import { defineStore } from 'pinia'
import { useSystemStore } from './system'
import axios from 'axios'

export const useEquipmentStore = defineStore('equipment', {
	state: () => ({
      workstations: [],
      equipment: [],
      pendingEquipment: {
         workstationID: 0,
         changed: false,
         equipment: []
      }
	}),
	getters: {
      scanners: state => {
         return state.equipment.filter( e => e.type == "Scanner")
      },
      lenses: state => {
         return state.equipment.filter( e => e.type == "Lens")
      },
      cameraBodies: state => {
         return state.equipment.filter( e => e.type == "CameraBody")
      },
      digitalBacks: state => {
         return state.equipment.filter( e => e.type == "DigitalBack")
      },
	},
	actions: {
      getEquipment( ) {
         const system = useSystemStore()
         axios.get( `/api/equipment` ).then(response => {
            this.workstations = response.data.workstations
            this.equipment = response.data.equipment
         }).catch( e => {
            system.setError(e)
         })
      },
      workstationSelected( wsID ) {
         this.pendingEquipment.workstationID = wsID
         this.pendingEquipment.changed = false
         let ws = this.workstations.find( ws => ws.id == wsID )
         this.pendingEquipment.equipment = ws.equipment.slice()
      },
      async addWorkstation( newName ) {
         var req = {name: newName}
         return axios.post( `/api/workstation`, req ).then((response) => {
            this.workstations.push( response.data )
         }).catch( e => {
            const system = useSystemStore()
            system.setError(e)
         })
      },
      async addEquipment( equipType, name, serialNumber) {
         var req = {type: equipType, name: name, serialNumber: serialNumber}
         return axios.post( `/api/equipment`, req ).then((response) => {
            this.equipment.push( response.data )
         }).catch( e => {
            const system = useSystemStore()
            system.setError(e)
         })
      },
      deactivateWorkstation( wsID ) {
         axios.post( `/api/workstation/${wsID}/update?status=1` ).then(() => {
            let tgtWS = this.workstations.find(ws => ws.id == wsID)
            if (tgtWS) {
               tgtWS.status = 1
            }
         }).catch( e => {
            const system = useSystemStore()
            system.setError(e)
         })
      },
      activateWorkstation( wsID ) {
         axios.post( `/api/workstation/${wsID}/update?status=0` ).then(() => {
            let tgtWS = this.workstations.find(ws => ws.id == wsID)
            if (tgtWS) {
               tgtWS.status = 0
            }
         }).catch( e => {
            const system = useSystemStore()
            system.setError(e)
         })
      },
      retireWorkstation( wsID ) {
         axios.post( `/api/workstation/${wsID}/update?status=2` ).then(() => {
            let wsIdx = this.workstations.findIndex(ws => ws.id == wsID)
            this.workstations.splice(wsIdx, 1)
         }).catch( e => {
            const system = useSystemStore()
            system.setError(e)
         })
      },
      async updateEquipment( equipID, newName, newSerial, newStatus ) {
         var req = {name: newName, serialNumber: newSerial, status: newStatus}
         return axios.post( `/api/equipment/${equipID}/update`, req ).then(() => {
            let owningWS = null
            let wsEquipIndex = -1
            this.workstations.some( ws => {
               wsEquipIndex = ws.equipment.findIndex( e => e.id == equipID)
               if ( wsEquipIndex > -1 ) {
                  owningWS = ws
               }
               return owningWS != null
            })
            if (newStatus == 2) {
               // retired; remove from equip list and workstation equipment
               let eIdx = this.equipment.findIndex(e => e.id == equipID)
               this.equipment.splice(eIdx, 1)
               if ( owningWS ) {
                  owningWS.equipment.splice(wsEquipIndex, 1)
               }
            } else {
               let tgtE = this.equipment.find(e => e.id == equipID)
               tgtE.status = newStatus
               tgtE.name = newName
               tgtE.serialNumber = newSerial
               if ( owningWS ) {
                  let tgtE = owningWS.equipment.find(e => e.id == equipID)
                  tgtE.status = newStatus
                  tgtE.name = newName
                  tgtE.serialNumber = newSerial
               }
            }
         }).catch( e => {
            console.log(e)
            const system = useSystemStore()
            system.setError(e)
         })
      },
      clearSetup() {
         this.pendingEquipment.changed = true
         this.pendingEquipment.equipment = []
      },
      async saveSetup() {
         var req = {setup: this.pendingEquipment.equipment}
         return axios.post( `/api/workstation/${this.pendingEquipment.workstationID}/setup`, req ).then((response) => {
            let wsIdx = this.workstations.findIndex( ws => ws.id == this.pendingEquipment.workstationID)
            this.workstations[wsIdx] = response.data.workstation
            this.equipment = response.data.equipment
            this.pendingEquipment.changed = false
         }).catch( e => {
            console.log(e)
            const system = useSystemStore()
            system.setError(e)
         })
      }
	}
})