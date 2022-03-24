import { data, fetchData } from './global'

const links = [
    {
        label: "Dashboard",
        link: "/dash",
        icon: "fa-solid fa-house",
    },
    {
        label: "Rooms",
        link: "/rooms",
        icon: "fa-solid fa-table-cells-large",
    },
    {
        label: "Profile",
        link: "/profile",
        icon: "fa-solid fa-user",
    },
]

main()

async function main() {
    const navbars = document.getElementsByTagName('nav')
    if (!navbars) return
    const navbar = navbars[0]
    navbar.className = 'nav closed'

    await fetchData()

    // Background
    const bg = document.createElement('div')
    bg.className = 'nav__bg threeDp'
    navbar.appendChild(bg)

    // Header
    const header = document.createElement('div')
    header.className = 'nav__header'
    navbar.appendChild(header)

    const headerAvatar = document.createElement('div')
    headerAvatar.className = 'nav__header__avatar'
    headerAvatar.style.backgroundImage = 'url(/api/user/avatar)'
    header.appendChild(headerAvatar)

    const headerTexts = document.createElement('div')
    headerTexts.className = 'nav__header__texts'
    header.appendChild(headerTexts)

    const headerTextFull = document.createElement('span')
    headerTextFull.className = 'nav__header__texts__full'
    headerTextFull.innerText = `${data.userData.forename} ${data.userData.surname}`
    headerTexts.appendChild(headerTextFull)

    const headerTextUser = document.createElement('span')
    headerTextUser.innerText = data.userData.username
    headerTexts.appendChild(headerTextUser)
}
