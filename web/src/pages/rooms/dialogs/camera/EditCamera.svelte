<script lang="ts">
    import Button,{ Icon,Label } from '@smui/button'
    import Dialog,{ Actions,Content,InitialFocus,Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { createEventDispatcher } from 'svelte'

    // Event dispatcher
    const dispatch = createEventDispatcher()
    function modifySelf() {
        dispatch('modify', null)
    }
    function deleteSelf() {
        dispatch('delete', null)
    }

    let deleteOpen = false
    export let open = false

    export let id: string
    export let name: string
    export let url: string

    let nameBefore: string
    let urlBefore: string
    let nameDirty = false

    $: nameDirty = name != nameBefore
    $: urlDirty = url != urlBefore

    export function show() {
        open = true
        nameBefore = name
        urlBefore = url
    }

    function cancel() {
        name = nameBefore
        url = urlBefore
    }
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Dialog
        bind:open={deleteOpen}
        slot="over"
        aria-labelledby="confirmation-title"
        aria-describedby="confirmation-content"
    >
        <Title id="confirmation-title">Confirm Deletion</Title>
        <Content id="confirmation-content">
            You are about to delete the camera '{name}'. This action is
            irreversible, do you want to proceed?
        </Content>
        <Actions>
            <Button on:click={deleteSelf}>
                <Label>Delete</Label>
            </Button>
            <Button use={[InitialFocus]}>
                <Label>Cancel</Label>
            </Button>
        </Actions>
    </Dialog>
    <Title id="title">Edit Camera <code>{id}</code></Title>
    <Content id="content">
        <Textfield bind:value={name} input$maxlength={30} label="Name" required>
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 30</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield bind:value={url} label="Url" type="url" />
        <div id="delete">
            <Button variant="outlined" on:click={() => (deleteOpen = true)}>
                <Icon class="material-icons">delete</Icon>
                <Label>Delete</Label>
            </Button>
        </div>
    </Content>
    <Actions>
        <Button on:click={cancel}>
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={!nameDirty && !urlDirty}
            use={[InitialFocus]}
            on:click={() => {
                nameBefore = name
                urlBefore = url
                modifySelf()
            }}
        >
            <Label>Modify</Label>
        </Button>
    </Actions>
</Dialog>

<style style="scss">
    code {
        background-color: var(--clr-height-0-3);
        padding: 0.1rem 0.5rem;
        border-radius: 0.3rem;
    }
    #delete {
        margin-top: 1rem;
    }
</style>
