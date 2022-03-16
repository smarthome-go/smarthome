addLoadEvent(function () {
  const navbars = document.getElementsByTagName("nav");
  if (!navbars) {
    return;
  }
  const navbar = navbars[0];
  navbar.className = "nav closed";

  //   Header
  const avatarImage = document.createElement("div");
  avatarImage.className = "nav__header__avatar__image__src";

  const avatarImageDiv = document.createElement("div");
  avatarImageDiv.className = "nav__header__avatar__image";
  avatarImageDiv.appendChild(avatarImage);

  const avatarMainSpan = document.createElement("span");
  avatarMainSpan.className = "nav__text nav__header__avatar__text__main";
  avatarMainSpan.innerText = "Mik Müller";

  const avatarSubSpan = document.createElement("span");
  avatarSubSpan.className = "nav__text nav__header__avatar__text__sub";
  avatarSubSpan.innerText = "Admin";

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
    navbar.classList.toggle("closed");
  };

  const header = document.createElement("header") as HTMLHtmlElement;
  header.className = "nav__header";
  header.appendChild(headerAvatarTextDiv);
  header.appendChild(toggle);

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

  navbar.appendChild(header);
  navbar.appendChild(menuBar);

  setTimeout(() => {
    navbar.style.transition = "var(--tran-03)";
  }, 100);
});
