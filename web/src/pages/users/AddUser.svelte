<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    export let open = false

    export let username = ''
    let usernameDirty = false
    export let password = ''
    let passwordDirty = false

    export function show() {
        open = true
        usernameDirty = false
        passwordDirty = false
    }

    export let onAdd = (_username: string, _password: string) => {}
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
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 20</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:dirty={passwordDirty}
            bind:value={password}
            label="Password"
            required
        />
    </Content>
    <Actions>
        <Button
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
