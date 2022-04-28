<script lang="ts">
    import Button,{ Icon,Label } from '@smui/button'
    import Dialog,{ Actions,Content,InitialFocus,Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { createSnackbar } from '../../../../global'
    import { loading,SwitchResponse } from '../../main'

    let deleteOpen = false
    let open = false

    export let switches: SwitchResponse[]
    export let id: string
    export let name: string
    export let watts: number

    let nameBefore: string
    let wattsBefore: number
    let nameDirty = false
    let wattsDirty = false

    $: nameDirty = name != nameBefore
    $: wattsDirty = watts != wattsBefore

    export function show() {
        open = true
        nameBefore = name
        wattsBefore = watts
    }

    function cancel() {
        name = nameBefore
        watts = wattsBefore
    }

    async function modifySwitch() {
        $loading = true
        try {
            const res = await (
                await fetch('/api/switch/modify', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ id, name, watts }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            nameBefore = name
            wattsBefore = watts
        } catch (err) {
            $createSnackbar(`Could not edit this switch: ${err}`)
        }
        $loading = false
    }

    async function deleteSwitch() {
        $loading = true
        try {
            const res = await (
                await fetch('/api/switch/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ id }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            switches = switches.filter((s) => s.id !== id)
            open = false
        } catch (err) {
            $createSnackbar(`Could not delete this switch: ${err}`)
        }
        $loading = false
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
            You are about to delete the switch '{name}'. This action is
            irreversible, do you want to proceed?
        </Content>
        <Actions>
            <Button on:click={deleteSwitch}>
                <Label>Delete</Label>
            </Button>
            <Button use={[InitialFocus]}>
                <Label>Cancel</Label>
            </Button>
        </Actions>
    </Dialog>
    <Title id="title">Edit Switch</Title>
    <Content id="content">
        <Textfield bind:value={name} input$maxlength={30} label="Name" required>
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 30</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield bind:value={watts} label="Watts" type="number" />
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
            disabled={!nameDirty && !wattsDirty}
            use={[InitialFocus]}
            on:click={modifySwitch}
        >
            <Label>Modify</Label>
        </Button>
    </Actions>
</Dialog>

<style style="scss">
    #delete {
        margin-top: 1rem;
    }
</style>
