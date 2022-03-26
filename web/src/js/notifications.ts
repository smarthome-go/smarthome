import { data, fetchNotifications, sleep } from './global'

export function createDrawer(): HTMLDivElement {
    const drawer = document.createElement('div')
    drawer.className = 'notifications hidden threeDp'

    const deleteButton = document.createElement('span')
    deleteButton.className = 'notifications__delete dummy'
    drawer.appendChild(deleteButton)

    for (let i = 0; i < data.notificationCount; i++) {
        const dummy = document.createElement('div')
        dummy.className = 'notification dummy oneDp'
        drawer.appendChild(dummy)

        const title = document.createElement('div')
        title.className = 'notification__title'
        dummy.appendChild(title)

        const desc1 = document.createElement('div')
        desc1.className = 'notification__desc'
        dummy.appendChild(desc1)

        const desc2 = document.createElement('div')
        desc2.className = 'notification__desc small'
        dummy.appendChild(desc2)

        const timestamp = document.createElement('div')
        timestamp.className = 'notification__time'
        dummy.appendChild(timestamp)
    }

    return drawer
}

export async function toggleDrawer(drawer: HTMLDivElement) {
    drawer.classList.toggle('hidden')
    if (data.notificationsLoaded) return

    data.notificationsLoaded = true

    data.notifications = await fetchNotifications()
    data.notificationCount = data.notifications.length

    // Remove dummy notifications
    drawer.innerHTML = ''

    if (data.notificationCount === 0) {
        addDoneMarker(drawer)
        return
    }

    const deleteButton = document.createElement('span')
    deleteButton.className = 'notifications__delete'
    deleteButton.innerText = 'DELETE ALL'
    deleteButton.onclick = async () => {
        const success = await deleteAll()
        if (!success) return
        // Copy current state
        const children = [...drawer.children as HTMLCollectionOf<HTMLElement>]
        for (const notification of children) {
            data.notifications.pop()
            updateIndicator()
            notification.style.transform = 'translateX(110%)'
            setTimeout(async () => drawer.firstElementChild?.remove(), 200)
            await sleep(50)
        }
        addDoneMarker(drawer)
    }
    drawer.appendChild(deleteButton)

    for (const notification of data.notifications) {
        const container = document.createElement('div')
        container.className = 'notification'
        drawer.appendChild(container)

        const line = document.createElement('div')
        line.className = 'notification__line'
        line.style.setProperty(
            '--clr-priority',
            notification.priority === 1 ? 'var(--clr-success)'
                : notification.priority === 2 ? 'var(--clr-warn)'
                    : 'var(--clr-error)'
        )
        container.appendChild(line)

        const deleteIcon = document.createElement('i')
        deleteIcon.className = 'notification__delete fa-solid fa-trash-can'
        deleteIcon.onclick = async () => {
            const success = await deleteNotification(notification.id)
            if (!success) return
            data.notifications.pop()
            updateIndicator()
            container.style.transform = 'translateX(110%)'
            await sleep(200)
            container.remove()
            if (data.notificationCount == 0) addDoneMarker(drawer)
        }
        container.appendChild(deleteIcon)

        const title = document.createElement('h3')
        title.innerText = notification.name
        container.appendChild(title)

        const description = document.createElement('p')
        description.innerText = notification.description
        container.appendChild(description)

        const timestamp = document.createElement('p')
        timestamp.className = 'notification__time'
        timestamp.innerText = notification.date
        container.appendChild(timestamp)
    }
}

function updateIndicator() {
    data.notificationCount = data.notifications.length
    const indicator = document.getElementsByClassName('nav__bell__icon__i__indicator')[0] as HTMLSpanElement
    indicator.innerHTML = `<span>${data.notificationCount}</span>`
    indicator.style.opacity = data.notificationCount === 0 ? '0' : '1'

    const bellText = document.getElementsByClassName('nav__bell__text')[0] as HTMLSpanElement
    bellText.innerText = data.notificationCount === 1 ? 'Notification' : 'Notifications'
}

function addDoneMarker(drawer: HTMLDivElement) {
    if (data.notificationCount !== 0 || data.notificationDoneMarkerAdded) return
    data.notificationDoneMarkerAdded = true
    drawer.innerHTML = ''

    const icon = document.createElement('i')
    icon.className = 'notifications__check fa-solid fa-check'
    drawer.appendChild(icon)

    const text = document.createElement('span')
    text.className = 'notifications__check-text'
    text.innerText = 'All caught up, no notifications'
    drawer.appendChild(text)

    setTimeout(() => {
        icon.style.opacity = '1'
        text.style.opacity = '1'
    }, 150)
}

async function deleteNotification(id: number): Promise<boolean> {
    return (await (await fetch('/api/user/notification/delete', {
        method: 'DELETE',
        body: JSON.stringify({ id: id }),
    })).json()).success
}

async function deleteAll(): Promise<boolean> {
    return (await (await fetch('/api/user/notification/delete/all', {
        method: 'DELETE',
    })).json()).success
}
