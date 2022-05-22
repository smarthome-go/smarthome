<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,InitialFocus,Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { createEventDispatcher } from 'svelte'
    export let open = false

    // Event dispatcher for deletion events
    const dispatch = createEventDispatcher()

    export let username = ''
    let usernameDirty = false
    export let password = ''
    let passwordDirty = false

    let confirmPassword = ''

    export function show() {
        open = true
        usernameDirty = false
        passwordDirty = false
    }

    // Will be used in order to show if a username is already taken
    export let blacklist: string[]

    let usernameInvalid = false
    $: usernameInvalid =
        (usernameDirty && username.length == 0) ||
        username.includes(' ') ||
        blacklist.includes(username)
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Add User</Title>
    <Content id="content">
        <Textfield
            bind:value={username}
            bind:dirty={usernameDirty}
            bind:invalid={usernameInvalid}
            input$maxlength={20}
            label="Username"
            required
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 20</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:value={password}
            bind:dirty={passwordDirty}
            label="Password"
            type="password"
            required
        />
        <br />
        <br />
        <Textfield
            bind:value={confirmPassword}
            invalid={password !== confirmPassword && passwordDirty}
            label="Repeat Password"
            type="password"
            required
        />
    </Content>
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={usernameInvalid ||
                password === '' ||
                password !== confirmPassword}
            use={[InitialFocus]}
            on:click={() => {
                dispatch('add', { username, password })
                username = ''
                password = ''
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>
