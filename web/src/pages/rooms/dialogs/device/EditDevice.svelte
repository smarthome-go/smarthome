<script lang="ts">
    import Button, { Icon, Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import Select, { Option } from '@smui/select'
    import { createEventDispatcher } from 'svelte'
    import { loading } from './main'
    import type { HydratedDeviceResponse } from '../../../../device';
    import { createDriverHMSID } from '../../../../driver';
    import { hmsEditorURLForId } from '../../../../urls';
    import DynamicConfigurator from '../../../../components/Homescript/DynamicConfigurator.svelte'
    import { createSnackbar, hasPermission } from '../../../../global';
    import type { Room } from '../../../../room';

    // Event dispatcher for deletion events
    const dispatch = createEventDispatcher()
    const deleteSelf = () => {
        dispatch('delete', null)
    }

    let deleteOpen = false
    let moveOpen = false
    let open = false

    export let rooms: Room[] = []
    export let data: HydratedDeviceResponse = null

    let dataBefore: HydratedDeviceResponse
    let moveTargetRoomId = ''

    export function show() {
        open = true
        dataBefore = structuredClone(data)
    }

    function cancel() {
        data = structuredClone(dataBefore)
        configuredChanged = false
    }

    let configuredChanged = false
    let configuredData = structuredClone(data.shallow.singletonJson)
    $: configuredChanged = (JSON.stringify(data.shallow.singletonJson) !== JSON.stringify(configuredData))
        || (JSON.stringify(data) !== JSON.stringify(dataBefore))


    function reactToOutput(modified: any) {
        if (!open) {
            return
        }

        configuredData = structuredClone(modified)
        configuredChanged = true
    }

    async function save() {
        dispatch('modify', data)
        await saveDeviceConfig()
    }

    async function saveDeviceConfig() {
        $loading = true

        try {
            let res = await fetch(
                '/api/devices/configure', {
                    method: "PUT",
                    body: JSON.stringify({
                        id: data.shallow.id,
                        data: configuredData
                    })
                }
            )

            if (res.status !== 200) {
                let msg  = await res.json()
                throw `${msg.message}: ${msg.error}`
            }

            configuredChanged = false
        } catch (err) {
            $createSnackbar(`Saving device configuration failed: ${err}`)
        }

        $loading = false
    }

    async function moveDevice() {
        if (!moveTargetRoomId) return

        $loading = true
        try {
            const res = await (
                await fetch('/api/devices/move', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        id: data.shallow.id,
                        roomId: moveTargetRoomId,
                    }),
                })
            ).json()

            if (!res.success) throw Error(res.error)

            moveOpen = false
            open = false
            dispatch('move', null)
        } catch (err) {
            $createSnackbar(`Failed to move device: ${err}`)
        }
        $loading = false
    }
</script>

<Dialog bind:open={moveOpen} aria-labelledby="move-title" aria-describedby="move-content">
    <Title id="move-title">Move to Other Room</Title>
    <Content id="move-content">
        <Select bind:value={moveTargetRoomId} label="Target Room" style="width: 100%;">
            {#each rooms.filter(r => r.data.id !== data.shallow.roomId) as room}
                <Option value={room.data.id}>{room.data.name}</Option>
            {/each}
        </Select>
    </Content>
    <Actions>
        <Button on:click={() => moveOpen = false}>
            <Label>Cancel</Label>
        </Button>
        <Button on:click={moveDevice} disabled={!moveTargetRoomId}>
            <Label>Move</Label>
        </Button>
    </Actions>
</Dialog>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Dialog
        bind:open={deleteOpen}
        slot="over"
        aria-labelledby="confirmation-title"
        aria-describedby="confirmation-content"
    >
        <Title id="confirmation-title">Confirm Deletion</Title>
        <Content id="confirmation-content">
            You are about to delete the device '{data.shallow.id}' (${data.shallow.name}}).
            This action is irreversible, do you want to proceed?
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
    <Title id="title">Edit Device <code>{data.shallow.id}</code></Title>
    <Content id="content">
        <Textfield bind:value={data.shallow.name} input$maxlength={30} label="Name" required>
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 30</CharacterCounter>
            </svelte:fragment>
        </Textfield>

        <div>
            <DynamicConfigurator
                bind:spec={data.extractions.config.info.config}
                on:change={ (e) => reactToOutput(e.detail) }
                bind:inputData={data.shallow.singletonJson}
                topLevelLabel={`Device Configuration`}
            />
        </div>

        {#if hasPermission('modifyServerConfig')}
            <br>
            <Button
                disabled={configuredChanged}
                variant="outlined"
                href={hmsEditorURLForId(createDriverHMSID(data.shallow.vendorId, data.shallow.modelId))}
            >
                <Icon class="material-icons">code</Icon>
                <Label>Edit Driver</Label>
            </Button>
        {/if}

        <div id="actions">
            <Button variant="outlined" on:click={() => (moveOpen = true)}>
                <Icon class="material-icons">swap_horiz</Icon>
                <Label>Move to Other Room</Label>
            </Button>

            <Button variant="outlined" on:click={() => (deleteOpen = true)}>
                <Icon class="material-icons">delete</Icon>
                <Label>Delete</Label>
            </Button>

            <Button disabled={!configuredChanged} variant="outlined" on:click={save}>
                <Icon class="material-icons">save</Icon>
                <Label>Save</Label>
            </Button>
        </div>
    </Content>
    <Actions>
        <Button on:click={cancel}>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>

<style style="scss">
    code {
        background-color: var(--clr-height-0-3);
        padding: 0.1rem 0.5rem;
        border-radius: 0.3rem;
    }
    #actions {
        margin-top: 1rem;
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
    }
</style>
