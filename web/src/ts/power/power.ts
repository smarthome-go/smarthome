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

function setCurrentRoom(room: Room, rooms: Room[]) {
  window.localStorage.setItem("current-room", room.id);
  const main = document.getElementById("current-room") as HTMLDivElement;
  const roomNavBar = document.getElementById("room-nav") as HTMLDivElement;
  while (main.firstChild) {
    main.removeChild(main.firstChild);
  }

  const roomDiv = createRoomDiv(room);
  main.appendChild(roomDiv);

  while (roomNavBar.firstChild) {
    roomNavBar.removeChild(roomNavBar.firstChild);
  }

  for (let roomItem of rooms) {
    const optionText = document.createElement("span") as HTMLSpanElement;
    optionText.innerText = roomItem.name;

    const option = document.createElement("div");
    option.className = "room-nav__option";
    if (roomItem.id == room.id) {
      option.classList.add("active");
    } else {
      option.onclick = () => {
        setCurrentRoom(roomItem, rooms);
      };
    }
    option.appendChild(optionText);

    roomNavBar.appendChild(option);
  }
}
addLoadEvent(async () => {
  const rooms = await getRooms();

  const main = document.getElementById("current-room") as HTMLDivElement;

  const currentRoom = window.localStorage.getItem("current-room");
  if (currentRoom != null) {
    for (let room of rooms) {
      if (room.id == currentRoom) {
        setCurrentRoom(room, rooms);
        return
      }
    }
    setCurrentRoom(rooms[0], rooms);
  } else {
    let roomWithMaxSwitches: Room = {
      id: "",
      description: "",
      name: "",
      switches: [],
    };
    for (let room of rooms) {
      if (room.switches.length > roomWithMaxSwitches.switches.length) {
        roomWithMaxSwitches = room;
      }
    }
    setCurrentRoom(roomWithMaxSwitches, rooms);
  }

  // const roomNavBar = document.getElementById("room-nav") as HTMLDivElement
  // for (let room of rooms) {
  //   const optionText = document.createElement("span") as HTMLSpanElement
  //   optionText.innerText = room.name

  //   const option = document.createElement("div")
  //   option.className = "room-nav__option"
  //   if (room.id == roomWithMaxSwitches.id) {
  //     option.classList.add("active")
  //   } else {
  //     option.onclick = () => {
  //       setCurrentRoom(room, rooms)
  //     }
  //   }
  //   option.appendChild(optionText)

  // roomNavBar.appendChild(option)
});
