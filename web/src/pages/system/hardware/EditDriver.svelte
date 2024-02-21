<script lang="ts">
    import Button, { Icon, Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import { createEventDispatcher } from 'svelte'
    import type { DriverData, FetchedDriver } from '../driver';
    import DynamicConfigurator from '../../../components/Homescript/DynamicConfigurator.svelte'
    import type { ConfigSpec } from 'src/driver';

    // Event dispatcher for deletion events
    const dispatch = createEventDispatcher()
    const deleteSelf = () => {
        open = false
        deleteOpen = false
        dispatch('delete', null)
    }

    let deleteOpen = false
    let open = false

    export let data: DriverData = {
            vendorId: "",
            modelId: "",
            name: "",
            version: "",
            homescriptCode: "",
    }

    let dataBefore: DriverData


    export function show() {
        open = true
        dataBefore = structuredClone(data)
    }

    function cancel() {
        data = structuredClone(dataBefore)
    }

    export let configSchema: ConfigSpec = null
    let configuredChanged = false
    export let dynamicConfig: any = null
    let configuredData: {} = null

    function reactToOutput(modified: any) {
        if (!open) {
            return
        }
        configuredData = structuredClone(modified)
        configuredChanged = true
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
            You are about to delete the device driver '{data.name}'.
            Deletion of a device driver can lead to unwanted consequences, such as the deletion of all devices which rely on that driver.
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
    <Title id="title">Edit Driver: <code>{data.vendorId}:{data.modelId}</code></Title>
    <Content id="content">
        <Textfield bind:value={data.name} input$maxlength={30} label="Name" required>
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 30</CharacterCounter>
            </svelte:fragment>
        </Textfield>

        <div>
            <br>
            <DynamicConfigurator
                bind:spec={configSchema}
                on:change={ (e) => reactToOutput(e.detail) }
                bind:inputData={dynamicConfig}
                topLevelLabel={`Driver Configuration`}
            />
        </div>

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
            disabled={false}
            use={[InitialFocus]}
            on:click={() =>
                dispatch('modify', {data, dynamic: configuredData})}
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
