<script lang="ts">
    import Button from '@smui/button/src/Button.svelte'
    import IconButton, { Icon } from '@smui/icon-button'
    import Tab, { Label } from '@smui/tab'
    import TabBar from '@smui/tab-bar'
    import { onMount } from 'svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar, hasPermission, sleep } from '../../global'
    import Page from '../../Page.svelte'
    import Camera from './Camera.svelte'
    import AddCamera from './dialogs/camera/AddCamera.svelte'
    import AddRoom from './dialogs/room/AddRoom.svelte'
    import EditRoom from './dialogs/room/EditRoom.svelte'
    import LocalSettings from './dialogs/room/LocalSettings.svelte'
    import AddDevice from './dialogs/device/AddDevice.svelte'
    import { loading, powerCamReloadEnabled } from './main'
    import Device from './Device.svelte'
    import type { Room } from '../../room';
    import type { CreateDeviceRequest } from '../../device';

    // If set to true, a camera-reload is triggered
    let reloadCameras = false

    // Specifies if all required data has been loaden
    // Used to hide the `no-xy` banners if the data is not loaded yet
    let loadedData = false

    // Whether the current-room dialog is open
    let editOpen = false
    let rooms: Room[]

    // Whether the local settings dialog is open
    let localSettingsOpen = false

    // Are bound backwards to pass the `open` event to the children
    let addRoomShow: () => void
    let addDeviceShow: () => void
    let addCameraShow: () => void

    let currentRoom: Room
    $: if (currentRoom !== undefined)
        window.localStorage.setItem('current_room', currentRoom.data.id)

    $: if (
        rooms !== undefined &&
        currentRoom !== undefined &&
        !rooms.find(r => r.data.id === currentRoom.data.id)
    )
        currentRoom = rooms.slice(-1)[0]

    // Determines if additional buttons for editing rooms should be visible
    let hasEditPermission: boolean
    let hasViewCamerasPermission: boolean
    onMount(async () => {
        hasEditPermission = await hasPermission('modifyRooms')
        hasViewCamerasPermission = await hasPermission('viewCameras')
    })

    // Fetches the available rooms.
    async function loadRooms(updateExisting = false) {
        $loading = true
        try {
            const personalOrAllSegment = (await hasPermission('modifyRooms')) ? 'all' : 'personal'
            const res = await ( await fetch( `/api/room/list/${personalOrAllSegment}`)).json()
            if (res.success === false) throw Error()
            if (updateExisting) {
                for (const room of rooms) {
                    room.devices = (res as Room[]).find(r => r.data.id === room.data.id).devices
                }
            } else {
                rooms = res
            }
            const roomId = window.localStorage.getItem('current_room')
            const room = roomId === null ? undefined : rooms.find(r => r.data.id === roomId)
            currentRoom = room === undefined ? rooms[0] : room
            loadedData = true
        } catch {
            $createSnackbar('Could not load rooms', [
                {
                    onClick: () => loadRooms(updateExisting),
                    text: 'retry',
                },
            ])
        }
        while (rooms === undefined) await sleep(10)
        $loading = false
    }

    // Adds a room
    async function addRoom(id: string, name: string, description: string) {
        $loading = true
        try {
            const res = await (
                await fetch(`/api/room/add`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ id, name }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            rooms = [
                ...rooms,
                {
                    data: {
                        id,
                        name,
                        description,
                    },
                    devices: [],
                    cameras: [],
                },
            ]
            rooms = rooms.sort((a, b) => a.data.name.localeCompare(b.data.name))
            await sleep(0) // Just for fixing js
            currentRoom = rooms[rooms.findIndex(r => r.data.id === id)]
        } catch (err) {
            $createSnackbar(`Failed to create room: ${err}`)
        }
        $loading = false
    }

    // Adds a device.
    async function addDevice(req: CreateDeviceRequest) {
        $loading = true

        try {
            req.roomId = currentRoom.data.id

            const res = await (
                await fetch('/api/devices/add', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(req),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            const currentRoomIndex = rooms.findIndex(r => r.data.id == currentRoom.data.id)

            // currentRoom.devices = [
            //     ...currentRoom.devices,
            //     {
            //         type: req.type,
            //         id: req.id,
            //         name: req.name,
            //         vendorId: req.vendorId,
            //         modelId: req.modelId,
            //         roomId: currentRoom.data.id,
            //         config: { info: null, capabilities: [] }, // TODO: get config from http-response
            //         singletonJson: null,
            //         dimmables: [],
            //         powerInformation: [],
            //
            //     },
            // ]

            // rooms[currentRoomIndex] = currentRoom
            await loadRooms()
        } catch (err) {
            $createSnackbar(`Could not create device: ${err}`)
        }
        $loading = false
    }

    // Adds a camera
    async function addCamera(id: string, name: string, url: string) {
        $loading = true
        try {
            const res = await (
                await fetch('/api/camera/add', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        id,
                        name,
                        url,
                        roomId: currentRoom.data.id,
                    }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            const currentRoomIndex = rooms.findIndex(r => r.data.id == currentRoom.data.id)
            currentRoom.cameras = [
                ...currentRoom.cameras,
                { id, name, url, roomId: currentRoom.data.id },
            ]
            rooms[currentRoomIndex] = currentRoom
        } catch (err) {
            $createSnackbar(`Could not create camera: ${err}`)
        }
        $loading = false
    }

    // Deletes a camera
    async function deleteCamera(id: string) {
        $loading = true
        try {
            const res = await (
                await fetch('/api/camera/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ id }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            currentRoom.cameras = currentRoom.cameras.filter(c => c.id !== id)
        } catch (err) {
            $createSnackbar(`Could not delete camera: ${err}`)
        }
        $loading = false
    }

    // Deletes a device.
    async function deleteDevice(id: string) {
        $loading = true
        try {
            const res = await (
                await fetch('/api/devices/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ id }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            currentRoom.devices = currentRoom.devices.filter(s => s.id !== id)
        } catch (err) {
            $createSnackbar(`Could not delete device: ${err}`)
        }
        $loading = false
    }

    async function modifyDevice(event) {
        const data = event.detail
        $loading = true
        try {
            const res = await (
                await fetch('/api/devices/modify', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            // TODO: make this generic: Would be reset on power change if not updated in `currentRoom`.
            let deviceInCurrentRoom = currentRoom.devices.find(s => s.id == data.id)
            deviceInCurrentRoom.name = data.name

            // TODO: add more?
        } catch (err) {
            $createSnackbar(`Could not edit device: ${err}`)
        }
        $loading = false
    }
</script>

<Page>
    <AddRoom blacklist={rooms} bind:show={addRoomShow} onAdd={addRoom} />
    <LocalSettings bind:open={localSettingsOpen} />
    {#if currentRoom !== undefined && hasEditPermission}
        <EditRoom
            bind:open={editOpen}
            bind:id={currentRoom.data.id}
            bind:name={currentRoom.data.name}
            bind:description={currentRoom.data.description}
            bind:rooms
        />
        <AddCamera cameras={currentRoom.cameras} bind:show={addCameraShow} onAdd={addCamera} />
        <AddDevice devices={currentRoom.devices} bind:show={addDeviceShow} on:add={(e) => addDevice(e.detail)} />
    {/if}
    <div id="tabs" class="mdc-elevation--z8">
        {#await loadRooms() then}
            <TabBar tabs={rooms} let:tab={room} bind:active={currentRoom} key={tab => tab.data.id}>
                <Tab tab={room} minWidth>
                    <Label>{room.data.name}</Label>
                </Tab>
            </TabBar>
        {/await}
        {#if currentRoom !== undefined}
            {#if hasEditPermission}
                <IconButton
                    class="material-icons"
                    title="Edit Current Room"
                    on:click={() => (editOpen = true)}>edit</IconButton
                >
            {/if}
            <IconButton
                class="material-icons"
                on:click={() => {
                    localSettingsOpen = true
                }}>settings</IconButton
            >
            <IconButton
                class="material-icons"
                on:click={() => {
                    loadRooms(true)
                }}>refresh</IconButton
            >
        {/if}
        {#if hasEditPermission}
            <IconButton class="material-icons" title="Add Room" on:click={addRoomShow}
                >add</IconButton
            >
        {/if}
        <Progress id="loader" bind:loading={$loading} />
    </div>

    <div id="content">
        <div id="devices" class="mdc-elevation--z1">
            {#if currentRoom == undefined && loadedData}
                <div id="no-rooms">
                    <i class="material-icons">no_meeting_room</i>
                    <h6>There are currently no rooms.</h6>
                    {#if hasEditPermission}
                        <div>
                            <Button variant="outlined" on:click={addRoomShow}>
                                <Label>Create Room</Label>
                                <Icon class="material-icons">add</Icon>
                            </Button>
                        </div>
                    {/if}
                </div>
            {:else}
                <!-- TODO: pack smaller devices into 'chunks' or 'groups' OR allow the user to pick an order? -->
                {#each currentRoom !== undefined ? currentRoom.devices : [] as device (device.id)}
                    <Device
                        bind:data={device}
                        on:delete={() => deleteDevice(device.id)}
                        on:modify={modifyDevice}
                        on:powerChange={() => (reloadCameras = $powerCamReloadEnabled)}
                        on:powerChangeDone={() => (reloadCameras = false)}
                    />
                {/each}
                {#if hasEditPermission}
                    <div id="add-device" class="switch mdc-elevation--z3">
                        <span>Add Device</span>
                        <IconButton class="material-icons" on:click={addDeviceShow}>add</IconButton>
                    </div>
                {:else if currentRoom !== undefined && currentRoom.devices.length == 0 && loadedData}
                    <div id="no-devices">
                        <i class="material-icons">power_off</i>
                        <h6>No Devices</h6>
                    </div>
                {/if}
            {/if}
        </div>
        <div
            id="cameras"
            class="mdc-elevation--z1"
            class:denied={!hasViewCamerasPermission && !hasEditPermission}
        >
            {#each currentRoom !== undefined ? currentRoom.cameras : [] as cam (cam.id)}
                <Camera
                    on:delete={() => deleteCamera(cam.id)}
                    id={cam.id}
                    name={cam.name}
                    url={cam.url}
                    reload={reloadCameras}
                />
            {/each}
            {#if hasEditPermission && currentRoom !== undefined}
                <div id="add-camera" class="switch mdc-elevation--z3">
                    <span>Add Camera</span>
                    <Button on:click={addCameraShow}>
                        <Label>Add</Label>
                        <Icon class="material-icons">add</Icon>
                    </Button>
                </div>
            {:else if currentRoom !== undefined && currentRoom.cameras.length == 0 && loadedData}
                <div id="no-cameras">
                    <i class="material-icons">videocam_off</i>
                    <h6>No Cameras</h6>
                </div>
            {/if}
        </div>
    </div>
</Page>

<style lang="scss">
    @use '../../mixins' as *;

    #no-rooms,
    #no-devices {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
        width: 100%;
        margin-top: 3rem;
        color: var(--clr-text-hint);

        i {
            font-size: 5rem;
        }
    }

    #no-cameras {
        // Similar to `#no-rooms`, but smaller
        display: flex;
        flex-direction: column;
        align-items: center;
        width: 100%;
        color: var(--clr-text-hint);

        i {
            font-size: 3rem;
        }

        h6 {
            margin: 0.3rem 0;
        }

        @include widescreen {
            margin-top: 1.5rem;
        }
    }

    #tabs {
        background-color: var(--clr-height-0-8);
        padding-right: 1rem;
        min-height: 48px;
        position: relative;
        display: flex;
        overflow-x: auto;

        & :global(#loader) {
            position: absolute;
            inset: 0;
            top: auto;
        }
    }
    #content {
        // On non-widescreen layouts, the room tabs might include a scrollbar which adds additional space
        min-height: calc(100vh - 63px);
        padding: 1rem 1.5rem;
        display: flex;
        gap: 1rem;
        flex-direction: column;
        box-sizing: border-box;

        @include widescreen {
            // On non-widescreen layouts, the room tabs might include a scrollbar which adds additional space
            // This space can be omitted on widescren
            min-height: calc(100vh - 48px);
            flex-direction: row;
        }
        @include mobile {
            min-height: calc(100vh - 48px - 3.5rem);
        }
    }
    #devices {
        background-color: var(--clr-height-0-1);
        padding: 1.5rem;
        border-radius: 0.4rem;
        display: flex;
        flex-wrap: wrap;
        gap: 1rem;
        align-content: flex-start;
        box-sizing: border-box;
        min-height: calc(100% - 16rem);
        flex-grow: 1;

        h6 {
            margin: 1rem 0;
        }

        @include widescreen {
            min-height: 100%;
        }

        @include mobile {
            flex-direction: column;
            flex-wrap: nowrap;
            align-content: unset;
            align-items: center;
        }
    }
    #cameras {
        background-color: var(--clr-height-0-1);
        height: 15rem;
        border-radius: 0.4rem;
        padding: 1.5rem;
        box-sizing: border-box;
        display: flex;
        gap: 1.5rem;
        overflow-x: auto;
        align-items: center;

        &.denied {
            opacity: 60%;
            pointer-events: none;
        }

        @include mobile {
            align-items: flex-start;
            justify-content: center;
            flex-direction: column;
            height: 100%;
        }

        @include widescreen {
            height: calc(100vh - 5rem);
            min-width: 21rem;
            flex-direction: column;
            overflow-y: auto;
            overflow-x: hidden;
            align-items: flex-start;
        }
    }
    #add-device,
    #add-camera {
        background-color: var(--clr-height-1-3);
        border-radius: 0.3rem;
        width: 17rem;
        height: 3.3rem;
        padding: 0.5rem;
        display: flex;
        justify-content: space-between;
        align-items: center;

        span {
            margin-left: 0.7rem;
            color: var(--clr-text-hint);
        }

        @include mobile {
            width: 90%;
            height: auto;
            flex-wrap: wrap;
        }
    }
    // Needed in order to account for special dimensions of the camera-layout
    #add-camera {
        flex-shrink: 0;
        height: 100%;
        width: auto;
        aspect-ratio: 16/9;
        padding: 1rem;
        box-sizing: border-box;
        position: relative;
        border-radius: 0.4rem;
        overflow: hidden;

        @include mobile {
            width: 100%;
            aspect-ratio: 16/9;
            box-sizing: border-box;
        }

        @include widescreen {
            width: 100%;
            height: auto;
        }
    }
</style>
