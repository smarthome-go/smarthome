// This file is being loaded by every route because it provides global data and utils

/*
GLOBAL DATA
*/

// Global datastore, it populated when the page loads
var data: Data = {
  userData: {
    firstname: "",
    primaryColor: "",
    surname: "",
    username: "",
  },
  notificationCount: 0,
  notifications: [],
};

// Global data interface
interface Data {
  notificationCount: number;
  notifications: Notification[];
  userData: UserData;
}

// User data fetched from the server
interface UserData {
  username: string;
  firstname: string;
  surname: string;
  primaryColor: string;
}

interface Notification {
  id: number;
  name: string;
  description: string;
  time: string;
}

// Fetches the data from the server when the page's navbar is loaded
async function loadData() {
  data.userData = await getUserData();
  data.notificationCount = await getNotificationCount();
  console.log("data: ", data);
}

async function getNotificationCount(): Promise<number> {
  const res = await fetch(`/api/user/notifications/count`);
  return (await res.json()).count;
}

async function getUserData(): Promise<UserData> {
  const res = await fetch(`/api/user/data`);
  return await res.json();
}

async function getNotifications() {
  const res = await fetch("/api/user/notifications");
  return await res.json()
}

/*
UTILS
*/
const sleep = (ms: number) => new Promise((res) => setTimeout(res, ms));

function addLoadEvent(func: () => void) {
  const oldOnLoad: any = window.onload;
  if (typeof window.onload != "function") {
    window.onload = func;
  } else {
    window.onload = function () {
      if (oldOnLoad) {
        oldOnLoad(undefined);
      }
      func();
    };
  }
}
