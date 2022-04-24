<script lang="ts">
    import Button,{ Icon,Label } from '@smui/button'
    import IconButton from '@smui/icon-button'
    import { onMount } from 'svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar } from '../../global'
    import Page from '../../Page.svelte'
    import AddUser from './dialogs/AddUser.svelte'
    import { allPermissions,loading,users } from './main'
    import User from './User.svelte'

    let addUserShow = () => {}

    export async function loadPermissions() {
        $loading = true
        try {
            $allPermissions = await (
                await fetch('/api/permissions/manage/list')
            ).json()
        } catch (err) {
            $createSnackbar(`Failed to load permissions: ${err}`)
        }
        $loading = false
    }

    async function loadUsers() {
        $loading = true
        try {
            const res = await (await fetch('/api/user/manage/list')).json()
            if (res.success !== undefined && !res.success)
                throw Error(res.error)
            $users = res.map((u) =>
                Object.create({
                    user: u,
                    permissions: [],
                    switchPermissions: [],
                })
            )
        } catch (err) {
            $createSnackbar(`Could not load users: ${err}`)
        }
        $loading = false
    }

    async function addUser(username: string, password: string) {
        $loading = true
        try {
            const res = await (
                await fetch('/api/user/manage/add', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, password }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            $users = [
                ...$users,
                {
                    user: {
                        darkTheme: true,
                        primaryColorDark: '#88FF70',
                        primaryColorLight: '#2E7D32',
                        schedulerEnabled: true,
                        forename: 'Forename',
                        surname: 'Surname',
                        username: username,
                    },
                    permissions: [],
                    switchPermissions: [],
                },
            ]
        } catch (err) {
            $createSnackbar(`Could not create user: ${err}`)
        }
        $loading = false
    }

    onMount(() => loadUsers())
</script>

<Page>
    <div id="container">
        <div id="header">
            <h6>User Management</h6>
            <div>
                <IconButton
                    title="Refresh"
                    class="material-icons"
                    on:click={loadUsers}>refresh</IconButton
                >
                <AddUser
                    onAdd={addUser}
                    bind:show={addUserShow}
                    blacklist={$users.map((u) => u.user.username)}
                />
                <Button on:click={addUserShow} variant="raised">
                    <Label>Add User</Label>
                    <Icon class="material-icons">person_add</Icon>
                </Button>
            </div>
        </div>
        <Progress id="loader" bind:loading={$loading} />
        <div id="users">
            {#each $users as user (user.user.username)}
                <div>
                    <User
                        {...user.user}
                        bind:permissions={user.permissions}
                        bind:switchPermissions={user.switchPermissions}
                    />
                </div>
            {/each}
        </div>
    </div>
</Page>

<style lang="scss">
    @use '../../mixins' as *;
    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: .1rem 1.3rem;
        box-sizing: border-box;
        background-color: var(--clr-height-1-2);

        @include mobile {
            flex-wrap: wrap;
        }

        h6 {
            margin: 0.5rem 0;
        }

        div {
            display: flex;
            align-items: center;
            gap: 1rem;
        }
    }
    #users {
        padding: 1.5rem;
        border-radius: 0.4rem;
        display: flex;
        flex-wrap: wrap;
        gap: 1rem;
        box-sizing: border-box;

        @include mobile {
            justify-content: center;
        }
    }
</style>
