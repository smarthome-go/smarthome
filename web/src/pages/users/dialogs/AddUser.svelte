<script lang="ts">
    import Button, { Label } from '@smui/button'
    import Dialog, { Actions, Content, Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    export let open = false

    export let username = ''
    let usernameDirty = false
    export let password = ''

    export function show() {
        open = true
        usernameDirty = false
    }

    // Will be used in order to show if a username is already taken
    export let blacklist: string[]

    export let onAdd = (_username: string, _password: string) => {}

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
            label="Username"
            input$maxlength={20}
            required
            bind:invalid={usernameInvalid}
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 20</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield bind:value={password} label="Password" />
    </Content>
    <Actions>
        <Button
            disabled={usernameInvalid}
            on:click={() => {
                onAdd(username, password)
                username = ''
                password = ''
            }}
        >
            <Label>Create</Label>
        </Button>
        <Button on:click={() => {}}>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>
