<script lang="ts">
    import Button,{ Icon,Label } from '@smui/button'
    import Dialog,{ Actions,Content,InitialFocus,Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { createSnackbar } from '../../../../global'
    import { loading,Room } from '../../main'

    export let open = false
    let deleteOpen = false

    export let id = ''
    export let name = ''
    export let description = ''

    let nameBefore = name
    let descriptionBefore = description

    export let rooms: Room[]

    async function modifyRoom() {
        $loading = true
        try {
            const res = await (
                await fetch('/api/room/modify', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        id,
                        name,
                        description,
                    }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            nameBefore = name
            descriptionBefore = description
            rooms[rooms.findIndex((r) => r.data.id === id)].data = {
                id,
                name,
                description,
            }
        } catch (err) {
            $createSnackbar(`Failed to modify room: ${err}`)
        }
        $loading = false
    }

    async function deleteRoom() {
        $loading = true
        try {
            const res = await (
                await fetch('/api/room/delete', {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ id }),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            rooms = rooms.filter((r) => r.data.id != id)
            open=false
        } catch (err) {
            $createSnackbar(`Failed to delete room: ${err}`)
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
            You are about to delete the room '{name}' and all its contents. This
            action is irreversible, do you want to proceed?
        </Content>
        <Actions>
            <Button on:click={deleteRoom}>
                <Label>Delete</Label>
            </Button>
            <Button use={[InitialFocus]}>
                <Label>Cancel</Label>
            </Button>
        </Actions>
    </Dialog>
    <Title id="title">Edit Room</Title>
    <Content id="content">
        <Textfield bind:value={name} input$maxlength={50} label="Name" required>
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 45</CharacterCounter>
            </svelte:fragment>
        </Textfield>
        <Textfield bind:value={description} label="Description" />
        <div id="delete">
            <Button variant='outlined'
                on:click={() => {
                    deleteOpen = true
                }}
            >
                <Icon class="material-icons">delete</Icon>
                <Label>Delete</Label>
            </Button>
        </div>
    </Content>
    <Actions>
        <Button
            on:click={() => {
                name = nameBefore
                description = descriptionBefore
            }}
        >
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={name == nameBefore && description == descriptionBefore}
            on:click={modifyRoom}
        >
            <Label>Modify</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    #delete {
        margin-top: 1rem;
    }
</style>
