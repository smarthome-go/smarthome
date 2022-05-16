<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{
    Actions,
    Content,
    Header,
    InitialFocus,
    Title
    } from '@smui/dialog'
    import FormField from '@smui/form-field'
    import IconButton from '@smui/icon-button'
    import Switch from '@smui/switch'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { onMount } from 'svelte'
    import ColorPicker from '../../../components/ColorPicker.svelte'
    import { createSnackbar,data } from '../../../global'
    import { loading,users } from './../main'

    // Dialog open / loading booleans
    export let open = false
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
    let schedulerEnabledBefore: boolean
    let darkThemeBefore: boolean

    const isCurrentUser = username == $data.userData.user.username

    // If the dialog edits the current user, some values can be changed directly in order to display a preview
    $: if (isCurrentUser) {
        $data.userData.user.darkTheme = darkTheme
        $data.userData.user.forename = forename
        $data.userData.user.surname = surname
        $data.userData.user.primaryColorDark = primaryColorDark
        $data.userData.user.primaryColorLight = primaryColorLight
    }

    // Variables that keep track of input change and valididy
    let forenameDirty = false
    let surnameDirty = false
    let forenameInvalid = false
    let surnameInvalid = false
    let primaryColorDarkDirty = false
    let primaryColorLightDirty = false
    // Update values reactively
    $: {
        forenameDirty = forename !== forenameBefore
        surnameDirty = surname !== surnameBefore

        primaryColorDarkDirty = primaryColorDark !== primaryColorDarkBefore
        primaryColorLightDirty = primaryColorLight !== primaryColorLightBefore

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
        schedulerEnabledBefore = schedulerEnabled
        darkThemeBefore = darkTheme
    }

    // Rolls back any changes
    function restoreChanges() {
        forename = forenameBefore
        surname = surnameBefore
        primaryColorDark = primaryColorDarkBefore
        primaryColorLight = primaryColorLightBefore
        schedulerEnabled = schedulerEnabledBefore
        darkTheme = darkThemeBefore
    }

    // Sends a delete request to the server
    async function deleteUser() {
        $loading = true
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
            $createSnackbar(`Could not delete user: ${err}`)
        }
        $loading = false
    }

    // Toggles the users scheduler
    async function setScheduler() {
        if (schedulerEnabled == schedulerEnabledBefore) return
        try {
            const res = await (
                await fetch('/api/scheduler/state/user', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        username,
                        enabled: schedulerEnabled,
                    }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
        } catch (err) {
            throw Error(err)
        }
    }

    // Toggles the users theme preference
    async function setTheme() {
        if (darkTheme == darkThemeBefore) return
        try {
            const res = await (
                await fetch('/api/user/settings/theme/user', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        username,
                        darkTheme,
                    }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
        } catch (err) {
            throw Error(err)
        }
    }

    // Sends a modification request to the server
    async function modify() {
        $loading = true
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
            await setScheduler()
            await setTheme()
            $createSnackbar(`Successfully modified user '${username}'`)
            updateBeforeValues()
        } catch (err) {
            $createSnackbar(`Failed to modify user data: ${err}`)
            restoreChanges()
        }
        $loading = false
    }
</script>

<Dialog bind:open fullscreen aria-labelledby="title" aria-describedby="content">
    <!-- Deletion confirmation dialog -->
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
            <Button on:click={deleteUser}>
                <Label>Delete</Label>
            </Button>
            <Button use={[InitialFocus]}>
                <Label>Cancel</Label>
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
        <div id="middle">
            <div id="toggles" class="mdc-elevation--z1">
                <!-- Boolean Toggles-->
                <Title>Toggles</Title>
                <FormField>
                    <Switch bind:checked={schedulerEnabled} />
                    <span id="toggles__scheduler__indicator" slot="label">
                        Scheduler {schedulerEnabled ? 'enabled' : 'disabled'}
                    </span>
                </FormField>
            </div>
            <div id="theming" class="mdc-elevation--z1">
                <!-- Theming Settings-->
                <Title>Theme</Title>
                <FormField>
                    <Switch bind:checked={darkTheme} />
                    <span slot="label"
                        >Dark Theme {darkTheme ? 'enabled' : 'disabled'}</span
                    >
                </FormField>
                <div id="primary-colors">
                    <div class="color">
                        <div
                            class="color-indicator"
                            style:background-color={primaryColorDark}
                        />
                        <!-- Primary Color Dark -->
                        <div>
                            <ColorPicker bind:value={primaryColorDark} />
                            <span>Dark</span>
                        </div>
                    </div>
                    <div class="color">
                        <div
                            class="color-indicator"
                            style:background-color={primaryColorLight}
                        />
                        <!-- Primary Color Light -->
                        <div>
                            <ColorPicker bind:value={primaryColorLight} />
                            <span>Light</span>
                        </div>
                    </div>
                </div>
            </div>
            <div id="danger">
                <!-- Dangerous actions: delete account -->
                <div id="danger__delete__user">
                    <div>
                        <Title>Delete User</Title>
                        <span class="--clr-text-hint"
                            >Erase all user data and delete account</span
                        >
                    </div>
                    <Button
                        variant="outlined"
                        on:click={() => {
                            deleteOpen = true
                        }}>delete</Button
                    >
                </div>
            </div>
        </div>
    </Content>
    <Actions>
        <!-- Only allow save if data has been changed -->
        <Button
            disabled={// Performs various integrity checks before sending data to the server for better UX
            (!forenameDirty &&
                !surnameDirty &&
                !primaryColorDarkDirty &&
                !primaryColorLightDirty &&
                schedulerEnabled == schedulerEnabledBefore &&
                darkTheme == darkThemeBefore) ||
                forenameInvalid ||
                surnameInvalid ||
                forename.length > 20 ||
                surname.length > 20}
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

<style lang="scss">
    @use '../../../mixins' as *;
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
    #primary-colors {
        display: flex;
        gap: 1rem;

        @include mobile {
            flex-direction: column;
        }

        .color {
            display: flex;
            align-items: center;
            gap: 0.2rem;
            background-color: var(--clr-height-1-3);
            padding: 1rem;
            border-radius: 0.3rem;

            div {
                display: flex;
                align-items: center;
                gap: 0.5rem;
            }
        }
    }
    .color-indicator {
        width: 2rem;
        height: 2rem;
        border-radius: 50%;

        @include not-widescreen {
            width: 1.2rem;
            height: 1.2rem;
        }
    }
    #middle {
        margin: 1rem 0;
        display: flex;
        gap: 1rem;
        flex-wrap: wrap;
        box-sizing: border-box;
    }
    #toggles,
    #theming,
    #danger {
        background-color: var(--clr-height-0-1);
        border-radius: 0.3rem;
        padding: 1rem;

        @include not-widescreen {
            width: 100%;
        }
    }
    #toggles__scheduler__indicator {
        display: block;
        min-width: 10rem;
    }
    #danger {
        &__delete__user {
            display: flex;
            gap: 2rem;
            align-items: flex-end;

            div {
                display: flex;
                flex-direction: column;
            }
        }
    }
</style>
