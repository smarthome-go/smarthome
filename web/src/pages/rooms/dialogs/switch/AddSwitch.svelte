<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,InitialFocus,Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import type { SwitchResponse } from '../../main'

    let open = false
    export let switches: SwitchResponse[] = []

    let id = ''
    let name = ''
    let watts = 0

    let idDirty = false
    let nameDirty = false

    export function show() {
        open = true
        id = ''
        name = ''
        watts = 0
        idDirty = false
        nameDirty = false
    }

    export let onAdd = (_id: string, _name: string, watts: number) => {}

    let idInvalid = false
    $: idInvalid =
        (idDirty && id === '') ||
        id.includes(' ') ||
        switches.find((s) => s.id === id) !== undefined
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Add Switch</Title>
    <Content id="content">
        <Textfield
            bind:value={id}
            bind:dirty={idDirty}
            bind:invalid={idInvalid}
            input$maxlength={20}
            label="Switch Id"
            required
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 20</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:value={name}
            bind:dirty={nameDirty}
            input$maxlength={30}
            label="Name"
            required
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 30</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield bind:value={watts} label="Watts" type="number" />
    </Content>
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={idInvalid || id === '' || name === ''}
            use={[InitialFocus]}
            on:click={() => {
                onAdd(id, name, watts)
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>
