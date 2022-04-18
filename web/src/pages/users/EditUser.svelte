<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Header,Title } from '@smui/dialog'
    import FormField from '@smui/form-field'
    import IconButton from '@smui/icon-button'
    import Paper from '@smui/paper'
    import Switch from '@smui/switch'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { createSnackbar,data } from '../../global'
    import { users } from './main'

    let open = false

    export let username = ''
    export let forename = ''
    export let surname = ''
    export let darkTheme: boolean
    export let automationEnabled: boolean
    export let permissions: string[]

    let deleteOpen = false

    let isSuspended = false
    $: {
        if (permissions !== null && permissions !== undefined)
            isSuspended = !permissions.includes('authentication')
    }

    $: {
        if (username == $data.userData.user.username)
            $data.userData.user.darkTheme = darkTheme
    }

    async function deleteUser() {
        try {
            const res = await (
                await fetch('/api/user/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            $createSnackbar(`Deleted user ${username}`)
            $users = $users.filter((u) => u.user.username !== username)
        } catch (err) {
            $createSnackbar(`Faield to delete user: ${err}`)
        }
    }

    // TODO: implement functions in GUI
    async function suspendUser() {
        try {
            const res = await (
                await fetch('/api/user/permissions/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        username,
                        permission: 'authentication',
                    }),
                })
            ).json()
            console.log(res)
            if (!res.success) throw Error(res.error)
            // Modify permission of user after succesfull suspension
            permissions = permissions.filter((p) => p != 'authentication')
        } catch (err) {
            $createSnackbar(`Failed to suspend user: ${err}`)
        }
    }

    async function activateUser() {
        try {
            const res = await (
                await fetch('/api/user/permissions/add', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        username,
                        permission: 'authentication',
                    }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            // Add authentication permission to local array
            permissions = [...permissions, 'authentication']
        } catch (err) {
            $createSnackbar(`Failed to activate user: ${err}`)
        }
    }
</script>

<Dialog bind:open fullscreen aria-labelledby="title" aria-describedby="content">
    <Dialog
        bind:open={deleteOpen}
        slot="over"
        aria-labelledby="confirmation-title"
        aria-describedby="confirmation-content"
    >
        <Title id="confirmation-title">Confirm Deletion</Title>
        <Content id="confirmation-content">
            You are about to delete the user '{username}'. This action is
            irreversible, do you want to proceed?
        </Content>
        <Actions>
            <Button>
                <Label>Cancel</Label>
            </Button>
            <Button on:click={deleteUser}>
                <Label>Delete</Label>
            </Button>
        </Actions>
    </Dialog>
    <Header>
        <Title id="title">Manage User</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
        <div id="profile">
            <img
                class="mdc-elevation--z3"
                src={`/api/user/avatar/user/${username}`}
                alt=""
            />
            <div>
                <h6>{forename} {surname}</h6>
                <span>{username}</span>
            </div>
        </div>
        <h6 id="edit">Edit</h6>
        <div id="names">
            <div>
                <!-- Forename -->
                <Textfield
                    helperLine$style="width: 100%;"
                    label="Forename"
                    input$maxlength={30}
                    bind:value={forename}
                >
                    <CharacterCounter slot="helper">0 / 30</CharacterCounter>
                </Textfield>
            </div>
            <div>
                <!-- Surname -->
                <Textfield
                    helperLine$style="width: 100%;"
                    label="Surname"
                    input$maxlength={30}
                    bind:value={surname}
                >
                    <CharacterCounter slot="helper">0 / 30</CharacterCounter>
                </Textfield>
            </div>
        </div>
        <div id="toggles" class="mdc-elevation--z1">
            <Paper color="primary" variant="outlined">
                <Title>Toggles</Title>
                <div id="toggle-content">
                    <FormField>
                        <Switch bind:checked={darkTheme} />
                        <span slot="label">Dark Theme</span>
                    </FormField>
                    <FormField>
                        <Switch bind:checked={automationEnabled} />
                        <span slot="label">Automation Enabled</span>
                    </FormField>
                </div>
            </Paper>
        </div>
        <div id="danger">
            <Paper variant="outlined">
                <Title>Dangerous Actions</Title>
                <div id="danger-buttons">
                    <div>
                        <Button
                            variant="outlined"
                            on:click={() => {
                                deleteOpen = true
                            }}>delete</Button
                        >
                        <span>Delete account</span>
                    </div>
                    <div>
                        <Button
                            variant="outlined"
                            on:click={isSuspended ? activateUser : suspendUser}
                        >
                            {isSuspended ? 'activate' : 'suspend'}</Button
                        >
                        <span
                            >{isSuspended
                                ? 'Activate account'
                                : 'Temporarily supend account'}</span
                        >
                    </div>
                </div>
            </Paper>
        </div>
    </Content>
    <Actions>
        <Button defaultAction>
            <Label>Done</Label>
        </Button>
        <Button>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>

<IconButton
    class="material-icons"
    on:click={async () => {
        open = true
    }}
    title="Manage">edit</IconButton
>

<style lang="scss">
    #names {
        display: flex;
        gap: 2rem;
    }
    #profile {
        display: flex;
        align-items: center;
        gap: 1rem;

        img {
            width: 5rem;
            height: 5rem;
            border-radius: 50%;
        }
    }
    h6 {
        margin: 0.5rem 0;
    }
    #edit {
        margin-top: 1rem;
    }

    #toggles {
        margin-top: 2rem;
        background-color: var(--clr-height-0-1);
    }
    #danger {
        margin-top: 1rem;
    }
    #danger-buttons {
        display: flex;
        gap: 3rem;
        margin-top: 0.7rem;

        div {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
    }
</style>
