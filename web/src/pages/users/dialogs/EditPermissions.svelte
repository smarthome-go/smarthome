<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Header,Title } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import Tab from '@smui/tab'
    import TabBar from '@smui/tab-bar'
    import Progress from '../../../components/Progress.svelte'
    import { createSnackbar } from '../../../global'
    import { fetchAllShallowDevices } from '../../../device'
    import {
    allCameras,
    allCamerasFetched,
    allPermissions,
    allDevices,
    allDevicesFetched,
    fetchAllCameras,
    fetchAllPermissions,
    fetchAllDevices
    } from '../main'
    import CameraPermissions from './CameraPermission.svelte'
    import Permission from './Permission.svelte'
    import DevicePermission from './DevicePermission.svelte'

    // Dialog open / loading booleans
    export let open = false
    export let currentMode: 'Permissions' | 'Device Permissions' | 'Camera Permissions'

    /**
     * Dynamic Content fetching
     * Only fetches content, such as user permissions when it is needed in order to make the website faster
     */

    // Keeps track of content and whether it has been fetched
    let permissionsFetched = false
    let devicePermissionsFetched = false
    let cameraPermissionsFetched = false

    // Exported user data
    export let username = ''
    export let permissions: string[] = []
    export let devicePermissions: string[] = []
    export let cameraPermissions: string[] = []

    $: {
        // Calls handleOpen when `open` or `currentMode` changes
        if (open && currentMode) handleOpen()
    }

    // Handles dynamic fetching of user data
    function handleOpen() {
        if ($allPermissions.length === 0) fetchAllPermissions()
        if (!$allDevicesFetched) fetchAllDevices()
        if (!$allCamerasFetched) fetchAllCameras()
        if (currentMode == 'Permissions' && !permissionsFetched)
            fetchUserPermissions()
        if (currentMode == 'Device Permissions' && !devicePermissionsFetched)
            fetchUserDevicePermissions()
        if (currentMode == 'Camera Permissions' && !cameraPermissionsFetched)
            fetchUserCameraPermissions()
    }

    /**
     * Functions for interacting with the backend
     * These functions communicate with the server in order to grant or deny permissions when they are updated
     */

    //  Retrieves the users personal permissions
    // TODO: remove this??? use store
    async function fetchUserPermissions() {
        try {
            const res = await (
                await fetch(`/api/user/permissions/list/user/${username}`)
            ).json()
            if (res.success !== undefined && !res.success)
                throw Error(res.error)
            permissionsFetched = true
            permissions = res
        } catch (err) {
            $createSnackbar(`Failed to load users permissions: ${err}`)
        }
    }

    // Retrieves the users personal switch permissions
    async function fetchUserDevicePermissions() {
        try {
            const res = await (
                await fetch(
                    `/api/user/permissions/device/list/user/${username}`
                )
            ).json()
            if (res.success !== undefined && !res.success)
                throw Error(res.error)

            devicePermissionsFetched = true
            devicePermissions = res
        } catch (err) {
            $createSnackbar(`Failed to load user switch permissions: ${err}`)
        }
    }

    // Retrieves the users personal camera permissions
    async function fetchUserCameraPermissions() {
        try {
            const res = await (
                await fetch(
                    `/api/user/permissions/camera/list/user/${username}`
                )
            ).json()
            if (res.success !== undefined && !res.success)
                throw Error(res.error)
            cameraPermissionsFetched = true
            cameraPermissions = res
        } catch (err) {
            $createSnackbar(`Failed to load user camera permissions: ${err}`)
        }
    }

    // Adds an arbitrary permission if it is valid and not held by the user
    async function grantPermission(permission: string) {
        try {
            const res = await (
                await fetch('/api/user/permissions/add', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, permission }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            permissions = [...permissions, permission]
        } catch (err) {
            $createSnackbar(`Failed to grant permission: ${err}`)
            throw Error()
        }
    }

    // Removes an arbitrary permission if it is valid and held by the user
    async function removePermission(permission: string) {
        try {
            const res = await (
                await fetch('/api/user/permissions/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, permission }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            permissions = permissions.filter((p) => p !== permission)
        } catch (err) {
            $createSnackbar(`Failed to remove permission: ${err}`)
            throw Error()
        }
    }

    // Adds an arbitrary device-permission if it is valid and not held by the user
    async function grantDevicePermission(permission: string) {
        try {
            const res = await (
                await fetch('/api/user/permissions/device/add', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, switch: permission }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            devicePermissions = [...devicePermissions, permission]
        } catch (err) {
            $createSnackbar(`Failed to grant device-permission: ${err}`)
            throw Error()
        }
    }

    // Removes an arbitrary device-permission if it is valid and held by the user
    async function removeDevicePermission(permission: string) {
        try {
            const res = await (
                await fetch('/api/user/permissions/device/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, switch: permission }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            devicePermissions = devicePermissions.filter(
                (s) => s !== permission
            )
        } catch (err) {
            $createSnackbar(`Failed to remove device-permission: ${err}`)
            throw Error()
        }
    }

    // Adds an arbitrary camera-permission if it is valid and not held by the user
    async function grantCameraPermission(id: string) {
        try {
            const res = await (
                await fetch('/api/user/permissions/camera/add', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, id }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            cameraPermissions = [...cameraPermissions, id]
        } catch (err) {
            $createSnackbar(`Failed to grant camera-permission: ${err}`)
            throw Error()
        }
    }

    // Removes an arbitrary camera-permission if it is valid and held by the user
    async function removeCameraPermission(id: string) {
        try {
            const res = await (
                await fetch('/api/user/permissions/camera/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, id }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            cameraPermissions = cameraPermissions.filter((s) => s !== id)
        } catch (err) {
            $createSnackbar(`Failed to remove camera-permission: ${err}`)
            throw Error()
        }
    }
</script>

<Dialog bind:open fullscreen aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">Manage User Permissions</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
        <div class="tabs">
            <TabBar
                tabs={permissions.includes('setPower')
                    ? permissions.includes('viewCameras')
                        ? [
                              'Permissions',
                              'Device Permissions',
                              'Camera Permissions',
                          ]
                        : ['Permissions', 'Device Permissions']
                    : permissions.includes('viewCameras')
                    ? ['Permissions', 'Camera Permissions']
                    : ['Permissions']}
                let:tab={mode}
                bind:active={currentMode}
                key={(tab) => tab}
            >
                <Tab tab={mode} minWidth>
                    <Label>{mode}</Label>
                </Tab>
            </TabBar>
            <IconButton
                title="Refresh"
                class="material-icons"
                on:click={() => {
                    switch (currentMode) {
                        case 'Permissions':
                            fetchUserPermissions()
                            break
                        case 'Device Permissions':
                            fetchUserDevicePermissions()
                            break
                        case 'Camera Permissions':
                            fetchUserCameraPermissions()
                            break
                    }
                }}>refresh</IconButton
            >
        </div>
        {#if currentMode === 'Permissions'}
            <div class="permissions">
                {#if $allPermissions.length == 0 || !permissionsFetched}
                    <div class="no-permissions">
                        <Progress type="circular" loading={true} />
                        <h6>Preparing editor...</h6>
                    </div>
                {:else}
                    {#each $allPermissions as permission (permission.permission)}
                        <Permission
                            description={permission.description}
                            name={permission.name}
                            permission={permission.permission}
                            grantFunc={grantPermission}
                            removeFunc={removePermission}
                            active={permissions.includes(permission.permission)}
                        />
                    {/each}
                {/if}
            </div>
        {:else if currentMode === 'Device Permissions'}
            <div class="switch-permissions">
                {#if !devicePermissionsFetched}
                    <div class="no-permissions">
                        <Progress type="circular" loading={true} />
                        <h6>Preparing editor...</h6>
                    </div>
                {:else}
                    {#each $allDevices as device (device.id)}
                        <DevicePermission
                            id={device.id}
                            name={device.name}
                            roomId={device.roomId}
                            active={devicePermissions.includes(device.id)}
                            grantFunc={grantDevicePermission}
                            removeFunc={removeDevicePermission}
                        />
                    {/each}
                {/if}
                {#if $allDevices.length === 0 && devicePermissionsFetched}
                    <div class="no-permissions">
                        <i class="material-icons">power_off</i>
                        <div class="bottom">
                            <h6>No Devices Available</h6>
                            <span
                                >You can create devices in the <a href="/rooms"
                                    >rooms</a
                                > section.</span
                            >
                        </div>
                    </div>
                {/if}
            </div>
        {:else if currentMode === 'Camera Permissions'}
            <div class="camera-permissions">
                {#if !cameraPermissionsFetched}
                    <div class="no-permissions">
                        <Progress type="circular" loading={true} />
                        <h6>Preparing editor...</h6>
                    </div>
                {:else}
                    {#each $allCameras as camera (camera.id)}
                        <CameraPermissions
                            id={camera.id}
                            name={camera.name}
                            active={cameraPermissions.includes(camera.id)}
                            grantFunc={grantCameraPermission}
                            removeFunc={removeCameraPermission}
                        />
                    {/each}
                {/if}
                {#if $allDevices.length === 0 && cameraPermissionsFetched}
                    <div class="no-permissions">
                        <i class="material-icons">videocam_off</i>
                        <div class="bottom">
                            <h6>No cameras available</h6>
                            <span
                                >You can create cameras in the <a href="/rooms"
                                    >rooms</a
                                > section.</span
                            >
                        </div>
                    </div>
                {/if}
            </div>
        {/if}
    </Content>
    <Actions>
        <Button>
            <Label>Done</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use '../../../mixins' as *;

    .permissions,
    .switch-permissions,
    .camera-permissions {
        display: flex;
        flex-wrap: wrap;
        align-content: flex-start;
        gap: 1rem;
        height: 60vh;
    }

    .tabs {
        margin-bottom: 1rem;
        display: flex;

        @include mobile {
            flex-wrap: wrap;
            gap: 1rem;
        }
    }

    .no-permissions {
        display: flex;
        flex-direction: column;
        align-items: center;
        width: 100%;
        gap: 1.5rem;
        margin-top: 7rem;
        color: var(--clr-text-hint);

        h6 {
            margin: 0.5rem 0rem;
        }
        a {
            color: var(--clr-primary);
        }
        i {
            font-size: 5rem;
        }

        .bottom {
            display: flex;
            flex-direction: column;
            align-items: center;
        }
    }
</style>
