<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,InitialFocus,Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { createEventDispatcher } from 'svelte'
    import type { addAutomation } from '../main'
    // Event dispatcher
    const dispatch = createEventDispatcher()

    let data: addAutomation = {
        days: [],
        description: '',
        enabled: true,
        homescriptId: '',
        hour: 0,
        minute: 0,
        name: '',
        timingMode: 'normal',
    }

    export let open = false
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content" fullscreen>
    <Title id="title">Add Automation</Title>
    <Content id="content">
        <Textfield
            bind:value={data.name}
            input$maxlength={1}
            label="Name"
            required
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 1</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield bind:value={data.description} label="Description" />
    </Content>
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={true}
            use={[InitialFocus]}
            on:click={() => {
                dispatch('add', data)
                // Reset values here
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
</style>
