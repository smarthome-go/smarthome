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

function setCurrentRoom(room: Room, rooms: Room[]) {
  window.localStorage.setItem("current-room", room.id);
  const main = document.getElementById("current-room") as HTMLDivElement;
  const switchDiv = document.getElementById("switches") as HTMLDivElement;
  const roomNavBar = document.getElementById("room-nav") as HTMLDivElement;

  const switches = document.getElementById("switches") as HTMLDivElement;
  while (switches.firstChild) {
    switches.removeChild(switches.firstChild)
  }

  for (let switchItem of room.switches) {
    const innerContainer = document.createElement("div");
    innerContainer.className = "switch-outer threeDp"

    let switchL;
    const switchE = document.createElement("input");
    switchE.type = "checkbox";
    switchE.checked = switchItem.powerOn;
    switchE.id = switchItem.id;

    const switchS = document.createElement("span");
    switchS.className = "slider round sixDp";

    switchL = document.createElement("label");
    switchL.className = "switch";
    switchL.appendChild(switchE);
    switchL.appendChild(switchS);

    const labelText = document.createElement("span");
    labelText.className = "outletName";

    innerContainer.appendChild(switchL);
    innerContainer.appendChild(labelText);

    switchE.addEventListener("change", async function () {
      console.log(switchE.checked);
    });

    const switchLabel = document.createElement("span")
    switchLabel.innerHTML = switchItem.name

    innerContainer.appendChild(switchLabel)
    switchDiv.appendChild(innerContainer)
  }

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
  if (rooms.length === 0) {
    alert("It seems like you don't have any switches configured.");
    return;
  }
  const main = document.getElementById("current-room") as HTMLDivElement;

  const currentRoom = window.localStorage.getItem("current-room");
  if (currentRoom != null) {
    for (let room of rooms) {
      if (room.id == currentRoom) {
        setCurrentRoom(room, rooms);
        return;
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
});
