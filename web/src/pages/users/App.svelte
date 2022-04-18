<script lang="ts">
    import Button,{ Icon,Label } from '@smui/button'
    import IconButton from '@smui/icon-button'
    import { onMount } from 'svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar } from '../../global'
    import Page from '../../Page.svelte'
    import AddUser from './AddUser.svelte'
    import { allPermissions,users } from './main'
    import User from './User.svelte'

    let addUserShow = () => {}

    let loading = false

    export async function loadPermissions() {
        loading = true
        try {
            $allPermissions = await (
                await fetch('/api/permissions/list')
            ).json()
        } catch (err) {
            $createSnackbar(`Failed to load permissions: ${err}`)
        }
        loading = false
    }

    async function loadUsers() {
        loading = true
        try {
            const res = await (await fetch('/api/user/list')).json()
            if (res.success !== undefined && !res.success)
                throw Error(res.error)
            $users = res
        } catch (err) {
            $createSnackbar(`Could not load users: ${err}`)
        }
        loading = false
    }

    async function addUser(username: string, password: string) {
        loading = true
        try {
            const res = await (
                await fetch('/api/user/add', {
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
                        primaryColorDark: '',
                        primaryColorLight: '',
                        schedulerEnabled: true,
                        forename: 'Forename',
                        surname: 'Surname',
                        username: username,
                    },
                    permissions: ["authentication"]
                }
            ]
        } catch (err) {
            $createSnackbar(`Could not create user: ${err}`)
        }
        loading = false
    }

    onMount(() => loadUsers())
</script>

<Page>
    <Progress id="loader" bind:loading />
    <div id="container">
        <div id="header">
            <h6>User Management</h6>
            <div>
                <IconButton
                    title="Refresh"
                    class="material-icons"
                    on:click={loadUsers}>refresh</IconButton
                >
                <AddUser onAdd={addUser} bind:show={addUserShow} />
                <Button on:click={addUserShow} variant="raised">
                    <Label>Add User</Label>
                    <Icon class="material-icons">person_add</Icon>
                </Button>
            </div>
        </div>
        <div id="users">
            {#each $users as user (user.user.username)}
                <div>
                    <User {...user.user} bind:permissions={user.permissions} />
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
        margin: 1rem 2rem;
        box-sizing: border-box;

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
    }
</style>
