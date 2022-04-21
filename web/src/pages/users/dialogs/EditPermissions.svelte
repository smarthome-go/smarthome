<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Header,Title } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import { createSnackbar } from '../../../global'
    import { allPermissions,fetchAllPermissions,loading } from '../main'
    import Permission from './Permission.svelte'

    // Dialog open / loading booleans
    export let open = false
    let confirmOpen = false

    // Exported user data
    export let username = ''
    export let forename = ''
    export let surname = ''
    export let permissions: string[] = []

    $: {
        if (open && $allPermissions.length === 0) {
            $loading = true
            fetchAllPermissions().then(() => ($loading = false))
        }
    }

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
</script>

<Dialog bind:open fullscreen aria-labelledby="title" aria-describedby="content">
    <Dialog
        bind:open={confirmOpen}
        slot="over"
        aria-labelledby="confirmation-title"
        aria-describedby="confirmation-content"
    >
        <Title id="confirmation-title">Confirm Action</Title>
        <Content id="confirmation-content">
            You are about to grant {forename}
            {surname} ({username}) a critical permission. Do you want to
            proceed?
        </Content>
        <Actions>
            <Button>
                <Label>Cancel</Label>
            </Button>
            <Button>
                <Label>Grant</Label>
            </Button>
        </Actions>
    </Dialog>
    <Header>
        <Title id="title">Manage User Permissions</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
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
    </Content>
    <Actions>
        <Button>
            <Label>Done</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use '../../../mixins' as *;
    #permissions {
        display: flex;
        flex-wrap: wrap;
        gap: 1rem;
    }
</style>
