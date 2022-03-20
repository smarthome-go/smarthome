addLoadEvent(async function () {
  const navbars = document.getElementsByTagName("nav");
  if (!navbars) {
    return;
  }
  const navbar = navbars[0];
  navbar.className = "nav closed threeDp";

  // Fetch user data before setting text content
  await loadData();

  //   Header
  const avatarImage = document.createElement("div");
  avatarImage.className = "nav__header__avatar__image__src";

  const avatarImageDiv = document.createElement("div");
  avatarImageDiv.className = "nav__header__avatar__image";
  avatarImageDiv.appendChild(avatarImage);

  const avatarMainSpan = document.createElement("span");
  avatarMainSpan.className = "nav__text nav__header__avatar__text__main";
  avatarMainSpan.innerText = `${data.userData.firstname} ${data.userData.surname}`;

  const avatarSubSpan = document.createElement("span");
  avatarSubSpan.className = "nav__text nav__header__avatar__text__sub";
  avatarSubSpan.innerText = data.userData.username;

  const avatarTextDiv = document.createElement("div") as HTMLDivElement;
  avatarTextDiv.className = "nav__text nav__header__avatar__text";
  avatarTextDiv.appendChild(avatarMainSpan);
  avatarTextDiv.appendChild(avatarSubSpan);

  const headerAvatarTextDiv = document.createElement("div");
  headerAvatarTextDiv.className = "nav__header__avatar";
  headerAvatarTextDiv.appendChild(avatarImageDiv);
  headerAvatarTextDiv.appendChild(avatarTextDiv);

  const toggleChevron = document.createElement("i");
  toggleChevron.className = "right fa-solid fa-chevron-right";

  const toggle = document.createElement("div");
  toggle.appendChild(toggleChevron);
  toggle.className = "nav__header__toggle";

  toggle.onclick = () => {
    notificationDrawer.classList.toggle("small");
    navbar.classList.toggle("closed");
  };

  const header = document.createElement("header") as HTMLHtmlElement;
  header.className = "nav__header";
  header.appendChild(headerAvatarTextDiv);
  header.appendChild(toggle);

  // Notification bell icon
  const bellIcon = document.createElement("i");
  bellIcon.className = "text fa-solid fa-bell nav__bell__icon";

  const notificationText = document.createElement("span");
  notificationText.className = "nav__bell__text nav__text";
  notificationText.innerText = `notification${
    data.notificationCount > 1 || data.notificationCount == 0 ? "s" : ""
  }`;

  const indicator = document.createElement("span");
  indicator.className = "nav__bell__indicator";
  if (data.notificationCount > 0) {
    indicator.style.opacity = "1";
  } else {
    indicator.style.opacity = "0";
  }
  indicator.innerText = `${data.notificationCount}`;

  const bellDiv = document.createElement("div");
  bellDiv.className = "nav__bell";
  bellDiv.appendChild(bellIcon);
  bellDiv.appendChild(indicator);
  bellDiv.appendChild(notificationText);

  bellDiv.onclick = () => {
    showNotificationDrawer();
  };

  //   Menu Bar
  const menuLinks = document.createElement("ul");
  menuLinks.className = "nav__menubar__menu__links";

  const menu = document.createElement("div");
  menu.className = "nav__menubar__menu";
  menu.appendChild(menuLinks);

  // Bottom Menu
  const bottomLinks = document.createElement("ul");
  bottomLinks.className = "nav__menubar__bottom nav__menubar__menu__links";

  const bottomMenu = document.createElement("div");
  bottomMenu.className = "nav__menubar__bottom";
  bottomMenu.appendChild(bottomLinks);

  const menuBar = document.createElement("div");
  menuBar.appendChild(menu);
  menuBar.appendChild(bottomMenu);
  menuBar.className = "nav__menubar";

  const links = [
    {
      label: "Home",
      link: "/dash",
      icon: "fa-solid fa-house",
    },
    {
      label: "Power",
      link: "/light",
      icon: "fa-solid fa-lightbulb",
    },
    {
      label: "Profile",
      link: "/profile",
      icon: "fa-solid fa-user",
    },
  ];

  for (let link of links) {
    const item = document.createElement("li");
    item.className = "nav__menubar__menu__links__item";

    const icon = document.createElement("i");
    icon.className = link.icon;

    const label = document.createElement("span");
    label.innerText = link.label;

    const itemA = document.createElement("a");
    itemA.href = link.link;
    itemA.appendChild(icon);
    itemA.appendChild(label);

    // Detect if the current url matches the current element
    if (window.location.href.split("/").pop() == link.link.split("/").pop()) {
      item.className += " active";
      itemA.href = "";
    }

    item.appendChild(itemA);
    menuLinks.appendChild(item);
  }

  //   Logout button
  const item = document.createElement("li");
  item.className = "nav__menubar__menu__links__item";

  const icon = document.createElement("i");
  icon.className = "fa-solid fa-arrow-right-from-bracket";

  const label = document.createElement("span");
  label.innerText = "logout";

  const itemA = document.createElement("a");
  itemA.href = "logout";
  itemA.appendChild(icon);
  itemA.appendChild(label);

  item.appendChild(itemA);

  bottomLinks.appendChild(item);

  // Notification Drawer
  const notificationContainer = document.createElement("div");
  notificationContainer.className = "notifications__container";

  // Add dummy elements (will later be removed)
  for (let i = 0; i < data.notificationCount; i++) {
    const dummyNotificationTitle = document.createElement("div");
    dummyNotificationTitle.className =
      "notifications__container__item__title dummy";
    const dummyNotificationDescription = document.createElement("div");
    dummyNotificationDescription.className =
      "notifications__container__item__description large dummy";
    const dummyNotificationDescription2 = document.createElement("div");
    dummyNotificationDescription2.className =
      "notifications__container__item__description dummy";

    const dummyNotification = document.createElement("div");
    dummyNotification.className = "notifications__container__item dummy oneDp";
    dummyNotification.appendChild(dummyNotificationTitle);
    dummyNotification.appendChild(dummyNotificationDescription);
    dummyNotification.appendChild(dummyNotificationDescription2);
    notificationContainer.appendChild(dummyNotification);
  }

  const notificationDrawer = document.createElement("div");
  notificationDrawer.className = "notifications hidden small";
  notificationDrawer.appendChild(notificationContainer);

  navbar.appendChild(header);
  navbar.appendChild(bellDiv);
  navbar.appendChild(menuBar);

  document.body.appendChild(notificationDrawer);

  setTimeout(() => {
    navbar.style.transition = "var(--tran-03)";
  }, 100);

  //   Detect screen size and open the navbar if it matches
  if (window.matchMedia("(min-width: 1500px)").matches) {
    navbar.classList.remove("closed");
    notificationDrawer.classList.remove("small");
  }

  window.onresize = () => {
    if (window.matchMedia("(min-width: 1500px)").matches) {
      navbar.classList.remove("closed");
      notificationDrawer.classList.remove("small");
    } else {
      navbar.classList.add("closed");
      notificationDrawer.classList.add("small");
    }
  };
});

async function showNotificationDrawer() {
  const drawer = document.getElementsByClassName(
    "notifications"
  )[0] as HTMLDivElement;
  drawer.classList.toggle("hidden");

  const container = document.getElementsByClassName(
    "notifications__container"
  )[0] as HTMLDivElement;

  // Check if the notifications are up to date
  if (!data.notificationsLoaded) {
    data.notificationsLoaded = true;
    // The notifications are not up to date and will be updated
    const notifications: Notification[] = await getNotifications();
    data.notifications = notifications;
    data.notificationCount = data.notifications.length;

    // Remove dummy notifications first
    while (container.firstChild) {
      container.removeChild(container.firstChild);
    }

    // If there are 0 notifications, add a indicator that nothing is there
    addDoneMarker();

    for (let notification of data.notifications) {
      const deleteIcon = document.createElement("i");
      deleteIcon.className =
        "notifications__container__item__delete fa-solid fa-trash-can";

      const priorityIcon = document.createElement("i");
      priorityIcon.className = "notifications__container__item__priority"

      let isDeleted = false
      deleteIcon.onclick = async () => {
        const success = await deleteNotification(notification.id);
        if (success && !isDeleted) {
          isDeleted = true
          data.notifications.pop();
          updateNotificationMarker();
          outer.style.minHeight = "0";
          outer.style.height = "0";
          outer.style.padding = "0";
          outer.style.opacity = "0";
          await sleep(200);
          outer.remove();

          if (data.notificationCount == 0) {
            addDoneMarker();
          }
        }
      };

      const title = document.createElement("h3");
      title.innerText = notification.name;

      const description = document.createElement("span");
      description.innerText = notification.description;

      const outer = document.createElement("div");
      outer.className = "notifications__container__item oneDp";

      outer.appendChild(deleteIcon);
      outer.appendChild(title);
      outer.appendChild(description);
      outer.appendChild(priorityIcon);

      if (notification.priority == 1) {
        outer.style.borderLeft = "solid 1px var(--clr-primary)";
        priorityIcon.className += " fa-solid fa-circle-info"
        priorityIcon.style.color = "var(--clr-primary)"
      } else if (notification.priority == 2) {
        outer.style.borderLeft = "solid 1px var(--clr-warn)";
        priorityIcon.className += " fa-solid fa-triangle-exclamation"
        priorityIcon.style.color = "var(--clr-warn)"
      } else if (notification.priority == 3) {
        outer.style.borderLeft = "solid 1px var(--clr-error)";
        priorityIcon.className += " fa-solid fa-circle-exclamation"
        priorityIcon.style.color = "var(--clr-error)"
      }

      container.appendChild(outer);
    }
  }
}

function updateNotificationMarker() {
  data.notificationCount = data.notifications.length;
  const notificationIndicator = document.getElementsByClassName(
    "nav__bell__indicator"
  )[0] as HTMLSpanElement;
  notificationIndicator.innerText = `${data.notificationCount}`;
  if (data.notificationCount > 0) {
    notificationIndicator.style.opacity = "1";
  } else {
    notificationIndicator.style.opacity = "0";
  }
}

async function deleteNotification(id: number): Promise<boolean> {
  const res = await fetch("/api/user/notification/delete", {
    method: "DELETE",
    body: JSON.stringify({
      id: id,
    }),
  });
  return (await res.json()).success;
}

async function deleteAllNotifications(): Promise<boolean> {
  const res = await fetch("/api/user/notification/delete/all", {
    method: "DELETE",
  });
  return (await res.json()).success;
}

function addDoneMarker() {
  if (data.notificationCount == 0 && !data.notificationDoneMarkerAdded) {
    data.notificationDoneMarkerAdded = true
    const checkmark = document.createElement("i");
    checkmark.className =
      "notifications__container__checkmark fa-solid fa-check";

    const checkmarkText = document.createElement("span");
    checkmarkText.className = "notifications__container__checkmark-text";
    checkmarkText.innerText = "All caught up, no notifications.";

    const container = document.getElementsByClassName(
      "notifications__container"
    )[0] as HTMLDivElement;

    container.appendChild(checkmark);
    container.appendChild(checkmarkText);
  }
}
