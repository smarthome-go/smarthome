<script lang="ts">
    import Button, { Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import Select, { Option } from '@smui/select'
    import Progress from '../../../../../src/components/Progress.svelte'
    import type { CreateDeviceRequest, DeviceResponse } from '../../main'
    import type { DriverData } from '../../../system/driver'
    import { drivers, driversLoaded, fetchHardwareNodes, loading, hardwareNodesLoaded, hardwareNodes, fetchDrivers } from './main'

    let selectedDriver: DriverData = {
        vendorId: "",
        modelId: "",
        name: "",
        version: "",
        homescriptCode: ""
    }

    let open = false
    export let devices: DeviceResponse[] = []

    let dataToAdd: CreateDeviceRequest = createEmptyDevice()
    let dataDirty = {
        id: false,
        name: false,
    }

    function createEmptyDevice(): CreateDeviceRequest {
        return {
            type: "INPUT",
            id: "",
            name: "",
            vendorId: "",
            modelId: "",
            roomId: ""
        }
    }

    export function show() {
        open = true

        if (!$hardwareNodesLoaded) {
            fetchHardwareNodes()
            fetchDrivers()
        }
    }

    export let onAdd: (
        _id: string,
        _name: string,
        _watts: number,
        _targetNodeUrl: string,
        _selectedDriverVendorId: string,
        _selectedDriverModelId: string,
    ) => Promise<void>

    let idInvalid = false
    $: idInvalid = (dataDirty.id && dataToAdd.id === '')
    || dataToAdd.id.includes(' ')
    || devices.find(d => d.id === dataToAdd.id) !== undefined
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Add Switch</Title>
    <Content id="content">
        {#if $driversLoaded}
            <Select bind:value={selectedDriver} label="Select Driver">
                {#each $drivers as driver}
                    <Option value={driver}>
                        {driver.driver.name} <code>{driver.driver.vendorId}: {driver.driver.modelId}</code>
                    </Option>
                {/each}
            </Select>
        {:else}
            <Progress bind:loading={$loading} />
        {/if}
        <br>
        <br>
        <Textfield
            bind:value={dataToAdd.id}
            bind:dirty={dataDirty.id}
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
            bind:value={dataToAdd.name}
            bind:dirty={dataDirty.name}
            input$maxlength={30}
            label="Name"
            required
        >
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 30</CharacterCounter>
            </svelte:fragment>
        </Textfield>
    </Content>
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>

        <!-- TODO: implement `create` button properly -->

        <!-- <Button -->
        <!--     disabled={idInvalid || dataToAdd.id.replaceAll(' ', '') === '' || dataToAdd.name.replaceAll(' ', '') === ''} -->
        <!--     use={[InitialFocus]} -->
        <!--     on:click={() => { -->
        <!--         onAdd( -->
        <!--             { -->
        <!--             id, name, -->
        <!--             watts, -->
        <!--             targetNodeUrl === 'none' ? null : targetNodeUrl, -->
        <!--             selectedDriver} -->
        <!--         ) -->
        <!--     }} -->
        <!-- > -->
        <!--     <Label>Create</Label> -->
        <!-- </Button> -->
    </Actions>
</Dialog>
