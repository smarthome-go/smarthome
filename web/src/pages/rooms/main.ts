import App from './App.svelte'

export interface Room {
  data: {
      id: string
      name: string
      description: string
  }
  switches: Switch[]
  cameras: Camera[]
}
export interface Switch {
  id: string
  name: string
  powerOn: boolean
  watts: number
}

export interface Camera {}

export default new App({
  target: document.body,
})
