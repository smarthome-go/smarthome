<script lang="ts">
    import Button, { Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import Select, { Option } from '@smui/select'
    import Progress from '../../../../../src/components/Progress.svelte'
    import type { CreateDeviceRequest, DeviceResponse, DeviceType } from '../../main'
    import type { DriverData, FetchedDriver } from '../../../system/driver'
    import { loading, drivers, driversLoaded, fetchDrivers } from './main'
    import { createEventDispatcher } from 'svelte';

    const dispatch = createEventDispatcher()

    let selectedType: DeviceType = 'INPUT'

    let selectedDriver: FetchedDriver = {
        isValid: false,
        configuration: null,
        info: null,
        validationErrors: null,
        driver: {
            vendorId: "",
            modelId: "",
            name: "",
            version: "",
            homescriptCode: ""
        }
    }
    $: console.log(selectedDriver)

    let open = false

    // Gets the existing devices from the outside to check for ID collisions.
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
            driverVendorId: "",
            driverModelId: "",
            roomId: ""
        }
    }

    export function show() {
        open = true

        if (!$driversLoaded) {
            fetchDrivers()
        }
    }

    let idInvalid = false
    $: idInvalid = (dataDirty.id && dataToAdd.id === '')
    || dataToAdd.id.includes(' ')
    || devices.find(d => d.id === dataToAdd.id) !== undefined

    function onAdd() {
        dataToAdd.driverVendorId = selectedDriver.driver.vendorId
        dataToAdd.driverModelId = selectedDriver.driver.modelId
        dataToAdd.type = selectedType

        console.dir(dataToAdd)

        dispatch('add', dataToAdd)
    }

    // TODO: validation code for IDs!
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Add Device</Title>
    <Content id="content">
        {#if $driversLoaded}
            <Select bind:value={selectedDriver} label="Select Driver">
                {#each $drivers as driver (`${driver.driver.vendorId}:${driver.driver.modelId}`)}
                    <Option value={driver}>
                        {driver.driver.name}
                    </Option>
                {/each}
            </Select>
        {:else}
            <Progress bind:loading={$loading} />
        {/if}
        <br>
        <Select bind:value={selectedType} label="Select Type">
            {#each ['INPUT', 'OUTPUT'] as option}
                <Option value={option}>
                    <code>{option}</code>
                </Option>
            {/each}
        </Select>
        <br>
        <Textfield
            bind:value={dataToAdd.id}
            bind:dirty={dataDirty.id}
            bind:invalid={idInvalid}
            input$maxlength={20}
            label="Device Id"
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

        <Button
            disabled={idInvalid || dataToAdd.id.replaceAll(' ', '') === '' || dataToAdd.name.replaceAll(' ', '') === ''}
            use={[InitialFocus]}
            on:click={onAdd}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>

<style lang='scss'>
    .driver {
        &__name {
            font-size: .8rem;
        }
    }
</style>
