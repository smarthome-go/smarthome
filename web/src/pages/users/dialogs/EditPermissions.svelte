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
    allPermissions,
    allSwitches,
    allSwitchesFetched,
    fetchAllPermissions,
    fetchAllSwitches
    } from '../main'
    import Permission from './Permission.svelte'
    import SwitchPermission from './SwitchPermission.svelte'

    // Dialog open / loading booleans
    export let open = false
    export let currentMode = 'Permissions'

    /**
     * Dynamic Content fetching
     * Only fetches content, such as user permissions when it is needed in order to make the website faster
     */

    // Keeps track of content and wether it has been fetched
    let permissionsFetched = false
    let switchPermissionsFetched = false
    let cameraPermissionsFetched = false

    // Exported user data
    export let username = ''
    export let permissions: string[] = []
    export let switchPermissions: string[] = []
    export let cameraPermissions: string[]

    $: {
        // Calls handleOpen when `open` or `currentMode` changes
        if (open && currentMode) handleOpen()
    }

    // Handles dynamic fetching of user data
    function handleOpen() {
        if ($allPermissions.length === 0) fetchAllPermissions()
        if (!$allSwitchesFetched) fetchAllSwitches()
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
    async function removeCameraPermission(permission: string) {
        try {
            const res = await (
                await fetch('/api/user/permissions/camera/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, switch: permission }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            cameraPermissions = cameraPermissions.filter(
                (s) => s !== permission
            )
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
        <div id="tabs">
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
                    currentMode === 'Permissions'
                        ? fetchUserPermissions()
                        : fetchUserSwitchPermissions()
                }}>refresh</IconButton
            >
        </div>
        {#if currentMode === 'Permissions'}
            {#if $allPermissions.length == 0 || !permissionsFetched}
                <Progress type="linear" loading={true} />
                <span>Preparing editor...</span>
            {/if}
            <div id="permissions">
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
            {#if !switchPermissionsFetched}
                <Progress type="linear" loading={true} />
                <span>Preparing editor...</span>
            {/if}
            <div id="switch-permissions">
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
            </div>
        {:else if currentMode === 'Camera Permissions'}
            {#if !cameraPermissionsFetched}
                <Progress type="linear" loading={true} />
                <span>Preparing editor...</span>
            {/if}
            <div id="switch-permissions">
                {#each $allCameras as camera (camera.id)}
                    <SwitchPermission
                        id={camera.id}
                        name={camera.name}
                        roomId={camera.roomId}
                        active={switchPermissions.includes(camera.id)}
                        grantFunc={grantSwitchPermission}
                        removeFunc={removeSwitchPermission}
                    />
                {/each}
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
    #permissions,
    #switch-permissions {
        display: flex;
        flex-wrap: wrap;
        gap: 1rem;
    }

    #tabs {
        margin-bottom: 1rem;
        display: flex;

        @include mobile {
            flex-wrap: wrap;
            gap: 1rem;
        }
    }
</style>
