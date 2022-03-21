interface Room {
  id: string;
  name: string;
  description: string;
  switches: Switch[];
}

interface Switch {
  id: string;
  name: string;
  roomId: string;
  powerOn: boolean;
}

async function getRooms(): Promise<Room[]> {
  const res = await fetch(`/api/room/list/personal`);
  const body: Room[] = await res.json();
  return body;
}

function createRoomDiv(room: Room): HTMLDivElement {
  const main = document.createElement("div");
  main.className = "room";

  return main;
}

function displayCurrent(rooms: Room[]) {
    const main = document.getElementsByTagName("main")[0] as HTMLDivElement;
    const currentRoom = window.location.href.split("/").pop()
    if (currentRoom != "default") {
        for (let room of rooms) {
            if (room.id == currentRoom) {
                const roomDiv = createRoomDiv(room);
                main.appendChild(roomDiv)
                console.log("Found current room");
                return
            }
        }
    }
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

    console.log("Using fallback room with maximum number of switches");
    const roomDiv = createRoomDiv(roomWithMaxSwitches);
    main.appendChild(roomDiv)
}

addLoadEvent(async () => {
  const rooms = await getRooms();
  displayCurrent(rooms)
});
