<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Header,Title } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import Tab from '@smui/tab'
    import TabBar from '@smui/tab-bar'
    import Progress from '../../../components/Progress.svelte'
    import { createSnackbar } from '../../../global'
    import {
    allCameras,
    allCamerasFetched,
    allPermissions,
    allSwitches,
    allSwitchesFetched,
    fetchAllCameras,
    fetchAllPermissions,
    fetchAllSwitches
    } from '../main'
    import CameraPermissions from './CameraPermissions.svelte'
    import Permission from './Permission.svelte'
    import SwitchPermission from './SwitchPermission.svelte'

    // Dialog open / loading booleans
    export let open = false
    export let currentMode = 'Permissions'

    /**
     * Dynamic Content fetching
     * Only fetches content, such as user permissions when it is needed in order to make the website faster
     */

    // Keeps track of content and whether it has been fetched
    let permissionsFetched = false
    let switchPermissionsFetched = false
    let cameraPermissionsFetched = false

    // Exported user data
    export let username = ''
    export let permissions: string[] = []
    export let switchPermissions: string[] = []
    export let cameraPermissions: string[] = []

    $: {
        // Calls handleOpen when `open` or `currentMode` changes
        if (open && currentMode) handleOpen()
    }

    // Handles dynamic fetching of user data
    function handleOpen() {
        if ($allPermissions.length === 0) fetchAllPermissions()
        if (!$allSwitchesFetched) fetchAllSwitches()
        if (!$allCamerasFetched) fetchAllCameras()
        if (currentMode == 'Permissions' && !permissionsFetched)
            fetchUserPermissions()
        if (currentMode == 'Switch Permissions' && !switchPermissionsFetched)
            fetchUserSwitchPermissions()
        if (currentMode == 'Camera Permissions' && !cameraPermissionsFetched)
            fetchUserCameraPermissions()
    }

    /**
     * Functions for interacting with the backend
     * These functions communicate with the server in order to grant or deny permissions when they are updated
     */

    //  Retrieves the users personal permissions
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
    async function fetchUserSwitchPermissions() {
        try {
            const res = await (
                await fetch(
                    `/api/user/permissions/switch/list/user/${username}`
                )
            ).json()
            if (res.success !== undefined && !res.success)
                throw Error(res.error)
            switchPermissionsFetched = true
            switchPermissions = res
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
            $createSnackbar(`Failed to grant permisssion: ${err}`)
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

    // Adds an arbitrary switch-permission if it is valid and not held by the user
    async function grantSwitchPermission(permission: string) {
        try {
            const res = await (
                await fetch('/api/user/permissions/switch/add', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, switch: permission }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            switchPermissions = [...switchPermissions, permission]
        } catch (err) {
            $createSnackbar(`Failed to grant switch-permisssion: ${err}`)
            throw Error()
        }
    }

    // Removes an arbitrary switch-permission if it is valid and held by the user
    async function removeSwitchPermission(permission: string) {
        try {
            const res = await (
                await fetch('/api/user/permissions//switch/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, switch: permission }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            switchPermissions = switchPermissions.filter(
                (s) => s !== permission
            )
        } catch (err) {
            $createSnackbar(`Failed to remove switch-permission: ${err}`)
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
            $createSnackbar(`Failed to grant camera-permisssion: ${err}`)
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
                              'Switch Permissions',
                              'Camera Permissions',
                          ]
                        : ['Permissions', 'Switch Permissions']
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
                        case 'Switch Permissions':
                            fetchUserSwitchPermissions()
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
                {/if}
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
            </div>
        {:else if currentMode === 'Switch Permissions'}
            <div class="switch-permissions">
                {#if !switchPermissionsFetched}
                    <div class="no-permissions">
                        <Progress type="circular" loading={true} />
                        <h6>Preparing editor...</h6>
                    </div>
                {/if}
                {#each $allSwitches as switchItem (switchItem.id)}
                    <SwitchPermission
                        id={switchItem.id}
                        name={switchItem.name}
                        roomId={switchItem.roomId}
                        active={switchPermissions.includes(switchItem.id)}
                        grantFunc={grantSwitchPermission}
                        removeFunc={removeSwitchPermission}
                    />
                {/each}
                {#if $allSwitches.length === 0 && switchPermissionsFetched}
                    <div class="no-permissions">
                        <i class="material-icons">power_off</i>
                        <div class="bottom">
                            <h6>No switches available</h6>
                            <span
                                >You can create switches in the <a href="/rooms"
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
                {/if}
                {#each $allCameras as camera (camera.id)}
                    <CameraPermissions
                        id={camera.id}
                        name={camera.name}
                        active={cameraPermissions.includes(camera.id)}
                        grantFunc={grantCameraPermission}
                        removeFunc={removeCameraPermission}
                    />
                {/each}
                {#if $allSwitches.length === 0 && cameraPermissionsFetched}
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
