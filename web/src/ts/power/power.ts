interface Room {
    id: string
    name: string
    description: string
    switches: Switch[]
}

interface Switch {
    id: string
    name: string
    roomId: string
    powerOn: boolean
}

interface PowerRequest {
    switch: string
    powerOn: boolean
}

async function getRooms(): Promise<Room[]> {
    const res = await fetch(`/api/room/list/personal`)
    const body: Room[] = await res.json()
    return body
}

async function setPower(switchId: string, powerOn: boolean): Promise<boolean> {
    const body: PowerRequest = {
        powerOn: powerOn,
        switch: switchId,
    }
    const res = await fetch(`/api/power/set`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    })
    const response = await res.json()
    return response.success
}

function setCurrentRoom(room: Room, rooms: Room[]) {
    window.localStorage.setItem("current-room", room.id)
    const main = document.getElementById("current-room") as HTMLDivElement
    const switchDiv = document.getElementById("switches") as HTMLDivElement
    const roomNavBar = document.getElementById("room-nav") as HTMLDivElement

    const switches = document.getElementById("switches") as HTMLDivElement
    while (switches.firstChild) {
        switches.removeChild(switches.firstChild)
    }

    for (let switchItem of room.switches) {
        const loader = document.createElement("span")
        loader.className = "loader disabled"

        const switchCheckBox = document.createElement("input")
        switchCheckBox.type = "checkbox"
        switchCheckBox.checked = switchItem.powerOn
        switchCheckBox.id = switchItem.id

        const slider = document.createElement("span")
        slider.className = "slider round sixDp"

        const switchLabel = document.createElement("label")
        switchLabel.className = "switch"
        switchLabel.appendChild(switchCheckBox)
        switchLabel.appendChild(slider)

        switchCheckBox.addEventListener("change", async function () {
            loader.classList.remove("disabled")
            console.log(switchCheckBox.checked)
            const success = await setPower(switchItem.id, switchCheckBox.checked)
            if (success) {
              loader.classList.add("disabled")
            } else {
              loader.classList.add("error")
              await sleep(1000)
              loader.classList.add("disabled")
              loader.classList.remove("error")
              // TODO: add a better popup
              alert(`An error occurred during set power of ${switchItem.name}`)
            }
        })

        const name = document.createElement("span")
        name.innerHTML = switchItem.name

        const innerContainer = document.createElement("div")
        innerContainer.className = "switch-outer threeDp"

        const leftSection = document.createElement("div")
        leftSection.className = "switch-outer__left"
        leftSection.appendChild(switchLabel)
        leftSection.appendChild(name)

        innerContainer.appendChild(leftSection)
        innerContainer.appendChild(loader)

        switchDiv.appendChild(innerContainer)
    }

    while (roomNavBar.firstChild) {
        roomNavBar.removeChild(roomNavBar.firstChild)
    }

    for (let roomItem of rooms) {
        const optionText = document.createElement("span") as HTMLSpanElement
        optionText.innerText = roomItem.name

        const option = document.createElement("div")
        option.className = "room-nav__option"
        if (roomItem.id == room.id) {
            option.classList.add("active")
        } else {
            option.onclick = () => {
                setCurrentRoom(roomItem, rooms)
            }
        }
        option.appendChild(optionText)

        roomNavBar.appendChild(option)
    }
}
addLoadEvent(async () => {
    const rooms = await getRooms()
    if (rooms.length === 0) {
        alert("It seems like you don't have any switches configured.")
        return
    }
    const main = document.getElementById("current-room") as HTMLDivElement

    const currentRoom = window.localStorage.getItem("current-room")
    if (currentRoom != null) {
        for (let room of rooms) {
            if (room.id == currentRoom) {
                setCurrentRoom(room, rooms)
                return
            }
        }
        setCurrentRoom(rooms[0], rooms)
    } else {
        let roomWithMaxSwitches: Room = {
            id: "",
            description: "",
            name: "",
            switches: [],
        }
        for (let room of rooms) {
            if (room.switches.length > roomWithMaxSwitches.switches.length) {
                roomWithMaxSwitches = room
            }
        }
        setCurrentRoom(roomWithMaxSwitches, rooms)
    }
})
