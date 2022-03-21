interface Switch {
    id: string
    name: string,
    roomId: string
}


async function getSwitches() {
 const res = await fetch(`/api/power/list/personal`)   

}