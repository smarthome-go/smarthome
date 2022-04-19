<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Header,Title } from '@smui/dialog'
    import FormField from '@smui/form-field'
    import IconButton from '@smui/icon-button'
    import Paper from '@smui/paper'
    import Switch from '@smui/switch'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { onMount } from 'svelte'
    import Progress from '../../components/Progress.svelte'
    import { createSnackbar,data } from '../../global'
    import { users } from './main'

    // Dialog open / loading booleans
    let open = false
    let loading = false
    let deleteOpen = false

    // Exported user data
    export let username = ''
    export let forename = ''
    export let surname = ''
    export let primaryColorDark
    export let primaryColorLight
    export let darkTheme: boolean
    export let schedulerEnabled: boolean

    // Values before modification
    let forenameBefore: string
    let surnameBefore: string
    let primaryColorDarkBefore: string
    let primaryColorLightBefore: string

    // If the dialog edits  current user, some values can be changed directly in order to display a preview
    $: {
        if (username == $data.userData.user.username) {
            $data.userData.user.darkTheme = darkTheme
            $data.userData.user.forename = forename
            $data.userData.user.surname = surname
        }
    }

    // Variables that keep track of input change and valididy
    let forenameDirty = false
    let surnameDirty = false
    let forenameInvalid = false
    let surnameInvalid = false
    // Update values reactively
    $: {
        forenameDirty = forename !== forenameBefore
        surnameDirty = surname !== surnameBefore

        forenameInvalid = forename.length == 0
        surnameInvalid = surname.length === 0
    }

    // Sets the values before modification to the currently visable values
    onMount(updateBeforeValues) // Saves the values initially
    function updateBeforeValues() {
        forenameBefore = forename
        surnameBefore = surname
        primaryColorDarkBefore = primaryColorDark
        primaryColorLightBefore = primaryColorLight
    }

    // Rolls back any changes
    function restoreChanges() {
        forename = forenameBefore
        surname = surnameBefore
        primaryColorDark = primaryColorDarkBefore
        primaryColorLight = primaryColorLightBefore
    }

    // Sends a delete request to the server
    async function deleteUser() {
        loading = true
        try {
            const res = await (
                await fetch('/api/user/manage/delete', {
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
        loading = false
    }

    // Sends a modification request to the server
    async function modify() {
        loading = true
        try {
            const res = await (
                await fetch('/api/user/manage/data/modify', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        username,
                        data: {
                            forename,
                            surname,
                            primaryColorDark,
                            primaryColorLight,
                        },
                    }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            $createSnackbar(`Successfully modified user '${username}'`)
            updateBeforeValues()
        } catch (err) {
            $createSnackbar(`Failed to modify user data: ${err}`)
            restoreChanges()
        }
        loading = false
    }
</script>

<Dialog bind:open fullscreen aria-labelledby="title" aria-describedby="content">
    <Progress id="loader" bind:loading />
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
            <!-- Profile Preview -->
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
                    input$maxlength={20}
                    bind:invalid={forenameInvalid}
                    bind:value={forename}
                >
                    <CharacterCounter slot="helper">0 / 20</CharacterCounter>
                </Textfield>
            </div>
            <div>
                <!-- Surname -->
                <Textfield
                    helperLine$style="width: 100%;"
                    label="Surname"
                    input$maxlength={20}
                    bind:invalid={surnameInvalid}
                    bind:value={surname}
                >
                    <CharacterCounter slot="helper">0 / 20</CharacterCounter>
                </Textfield>
            </div>
        </div>
        <div id="toggles" class="mdc-elevation--z1">
            <!-- Boolean Toggles-->
            <Paper color="primary" variant="outlined">
                <Title>Toggles</Title>
                <div id="toggle-content">
                    <FormField>
                        <Switch bind:checked={darkTheme} />
                        <span slot="label">Dark Theme</span>
                    </FormField>
                    <FormField>
                        <Switch bind:checked={schedulerEnabled} />
                        <span slot="label">Scheduler Enabled</span>
                    </FormField>
                </div>
            </Paper>
        </div>
        <div id="danger">
            <!-- Dangerous actions: delete account -->
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
                </div>
            </Paper>
        </div>
    </Content>
    <Actions>
        <!-- Only allow save if data has been changed -->
        <Button
            disabled={(!forenameDirty && !surnameDirty) ||
                forenameInvalid ||
                surnameInvalid}
            defaultAction
            on:click={modify}
        >
            <Label>Save</Label>
        </Button>
        <!-- Restore changes if the user cancels the action -->
        <Button on:click={restoreChanges}>
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
