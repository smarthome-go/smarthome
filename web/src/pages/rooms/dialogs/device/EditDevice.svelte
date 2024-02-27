<script lang="ts">
    import Button, { Icon, Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { createEventDispatcher } from 'svelte'
    import { loading } from './main'
    import type { DeviceResponse } from '../../main';
    import DynamicConfigurator from '../../../../components/Homescript/DynamicConfigurator.svelte'
    import { createSnackbar, hasPermission } from '../../../../global';

    // Event dispatcher for deletion events
    const dispatch = createEventDispatcher()
    const deleteSelf = () => {
        dispatch('delete', null)
    }

    let deleteOpen = false
    let open = false

    export let data: DeviceResponse = null

    let dataBefore: DeviceResponse

    export function show() {
        open = true
        dataBefore = structuredClone(data)
    }

    function cancel() {
        data = structuredClone(dataBefore)
        configuredChanged = false
    }

    let configuredChanged = false
    let configuredData = structuredClone(data.singletonJson)
    $: configuredChanged = (JSON.stringify(data.singletonJson) !== JSON.stringify(configuredData))
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
                        id: data.id,
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
            You are about to delete the device '{data.id}' (${data.name}}).
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
    <Title id="title">Edit Device <code>{data.id}</code></Title>
    <Content id="content">
        <Textfield bind:value={data.name} input$maxlength={30} label="Name" required>
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 30</CharacterCounter>
            </svelte:fragment>
        </Textfield>

        <div>
            <DynamicConfigurator
                bind:spec={data.config.info.config}
                on:change={ (e) => reactToOutput(e.detail) }
                bind:inputData={data.singletonJson}
                topLevelLabel={`Device Configuration`}
            />
        </div>

        {#if hasPermission('modifyServerConfig')}
            <br>
            <Button
                disabled={configuredChanged}
                variant="outlined"
                href={hmsEditorURLForId(createDriverHMSID(data.vendorId, data.modelId))}
            >
                <Icon class="material-icons">code</Icon>
                <Label>Edit Driver</Label>
            </Button>
        {/if}

        <div id="delete">
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
    #delete {
        margin-top: 1rem;
    }
</style>
