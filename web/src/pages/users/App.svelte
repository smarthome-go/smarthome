<script lang="ts">
    import Button,{ Icon,Label } from '@smui/button'
    import { onMount } from 'svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar } from '../../global'
    import Page from '../../Page.svelte'
    import AddUser from './AddUser.svelte'
    import User from './User.svelte'

    let addUserShow = () => {}

    interface User {
        username: string
        forename: string
        surname: string
        primaryColorDark: string
        primaryColorLight: string
        schedulerEnabled: boolean
        darkTheme: boolean
    }

    let users: User[] = []
    let loading = false

    async function loadUsers() {
        loading = true
        try {
            const res = await (await fetch('/api/user/list')).json()
            if (res.success !== undefined && !res.success)
                throw Error(res.error)
            users = res
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
            users = [
                ...users,
                {
                    darkTheme: true,
                    primaryColorDark: '',
                    primaryColorLight: '',
                    schedulerEnabled: true,
                    forename: 'Forename',
                    surname: 'Surname',
                    username: username,
                },
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
            <AddUser onAdd={addUser} bind:show={addUserShow} />
            <Button on:click={addUserShow} variant="raised">
                <Label>Add User</Label>
                <Icon class="material-icons">person_add</Icon>
            </Button>
        </div>
        <div id="users">
            {#each users as user (user.username)}
            <div>
                <User {...user} />
            </div>
            {/each}
        </div>
    </div>
</Page>

<style lang="scss">
    @use '../../mixins' as *;
    #container {
        margin: 1rem 1.5rem;
        border-radius: 0.4rem;
        padding: 1.5rem;

        @include widescreen {
            padding: 1.5rem 20%;
        }
    }
    #users {
        div:not(:last-child) {
            margin-bottom: 1.3rem;
        }
    }
    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }
    h6 {
        margin: 1rem;
    }
</style>
