package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type staffMember struct {
	ID          uint   `json:"id"`
	ComputingID string `json:"computingID"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Role        uint   `json:"role"`
	Email       string `json:"email"`
}

func (sm *staffMember) roleString() string {
	roles := []string{"admin", "supervisor", "student", "viewer"}
	if sm.Role > uint(len(roles)-1) {
		return "viewer"
	}
	return roles[sm.Role]
}

type jwtClaims struct {
	UserID    uint   `json:"userID"`
	ComputeID string `json:"computeID"`
	Role      string `json:"role"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	jwt.RegisteredClaims
}

func (svc *serviceContext) authenticate(c *gin.Context) {
	log.Printf("Checking authentication headers...")
	// log.Printf("Dump all request headers ==================================")
	// for name, values := range c.Request.Header {
	// 	for _, value := range values {
	// 		log.Printf("%s=%s\n", name, value)
	// 	}
	// }
	// log.Printf("END header dump ===========================================")

	computingID := c.GetHeader("remote_user")
	if svc.DevAuthUser != "" {
		computingID = svc.DevAuthUser
		log.Printf("Using dev auth user ID: %s", computingID)
	}
	if computingID == "" {
		log.Printf("ERROR: Expected auth header not present in request. Not authorized.")
		c.Redirect(http.StatusFound, "/forbidden")
		return
	}

	// if not in dev mode, ensure membership in DPG group
	if svc.DevAuthUser == "" {
		// Membership format: cn=group_name1;cn=group_name2;...
		membershipStr := c.GetHeader("member")
		if !strings.Contains(membershipStr, "lb-digiserv") {
			log.Printf("ERROR: %s is not part of digiserv", computingID)
			c.Redirect(http.StatusFound, "/forbidden")
			return
		}
	}

	log.Printf("INFO: lookup staff member %s", computingID)
	respBytes, reqErr := svc.getRequest(fmt.Sprintf("%s/staff/lookup?cid=%s", svc.TrackSys.API, computingID))
	if reqErr != nil {
		log.Printf("ERROR: could not find staff member %s: %s", computingID, reqErr.Message)
		c.Redirect(http.StatusFound, "/forbidden")
		return
	}

	var sm staffMember
	if err := json.Unmarshal(respBytes, &sm); err != nil {
		log.Printf("ERROR: could not parse staff member %s reaponse: %s", computingID, err.Error())
		c.Redirect(http.StatusFound, "/forbidden")
		return
	}

	log.Printf("Generate JWT for %s", computingID)
	expirationTime := time.Now().Add(8 * time.Hour)
	claims := jwtClaims{
		UserID:    sm.ID,
		ComputeID: computingID,
		FirstName: sm.FirstName,
		LastName:  sm.LastName,
		Role:      sm.roleString(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "dpgimaging",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, jwtErr := token.SignedString([]byte(svc.JWTKey))
	if jwtErr != nil {
		log.Printf("ERROR: unable to generate JWT for %s: %s", computingID, jwtErr.Error())
		c.Redirect(http.StatusFound, "/forbidden")
		return
	}

	// Set auth info in a cookie the client can read and pass along in future requests
	c.SetCookie("dpg_jwt", signedStr, 10, "/", "", false, false)
	c.SetSameSite(http.SameSiteLaxMode)
	c.Redirect(http.StatusFound, "/granted")
}

// AuthMiddleware is middleware that checks for a user auth token in the
// Authorization header. For now, it does nothing but ensure token presence.
func (svc *serviceContext) authMiddleware(c *gin.Context) {
	log.Printf("Authorize access to %s", c.Request.URL)
	tokenStr, err := getBearerToken(c.Request.Header.Get("Authorization"))
	if err != nil {
		log.Printf("Authentication failed: [%s]", err.Error())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if tokenStr == "undefined" {
		log.Printf("Authentication failed; bearer token is undefined")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("Validating JWT auth token...")
	jwtClaims := jwtClaims{}
	_, jwtErr := jwt.ParseWithClaims(tokenStr, &jwtClaims, func(token *jwt.Token) (any, error) {
		return []byte(svc.JWTKey), nil
	})
	if jwtErr != nil {
		log.Printf("Authentication failed; token validation failed: %+v", jwtErr)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("got valid bearer token: [%s] for %s", tokenStr, jwtClaims.ComputeID)
	c.Set("jwt", tokenStr)
	c.Set("claims", jwtClaims)
	c.Next()
}

func getJWTClaims(c *gin.Context) *jwtClaims {
	claims, signedIn := c.Get("claims")
	if !signedIn {
		return nil
	}
	jwtClaims, ok := claims.(jwtClaims)
	if !ok {
		return nil
	}
	log.Printf("INFO: pulled claims from context [%+v]", jwtClaims)
	return &jwtClaims
}

// getBearerToken is a helper to extract the token from headers
func getBearerToken(authorization string) (string, error) {
	components := strings.Split(strings.Join(strings.Fields(authorization), " "), " ")

	// must have two components, the first of which is "Bearer", and the second a non-empty token
	if len(components) != 2 || components[0] != "Bearer" || components[1] == "" {
		return "", fmt.Errorf("invalid authorization header: [%s]", authorization)
	}

	return components[1], nil
}
