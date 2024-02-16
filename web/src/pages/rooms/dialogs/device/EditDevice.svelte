<script lang="ts">
    import Button, { Icon, Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import Select, { Option } from '@smui/select'
    import Progress from '../../../../../src/components/Progress.svelte'
    import { createEventDispatcher } from 'svelte'
    import { fetchHardwareNodes, loading, hardwareNodesLoaded, hardwareNodes } from './main'
    import type { DeviceResponse } from '../../main';
    import DynamicConfigurator from '../../../../components/Homescript/DynamicConfigurator.svelte'
    import { createSnackbar } from '../../../../global';

    // Event dispatcher for deletion events
    const dispatch = createEventDispatcher()
    const deleteSelf = () => {
        dispatch('delete', null)
    }

    let deleteOpen = false
    let open = false

    export let data: DeviceResponse = null

    let dataBefore: DeviceResponse

    // $: nameDirty = name != nameBefore
    // $: wattsDirty = watts != wattsBefore
    // $: targetNodeDirty = targetNode != targetNodeBefore

    export function show() {
        open = true
        // nameBefore = name
        // wattsBefore = watts
        // targetNodeBefore = targetNode

        dataBefore = structuredClone(data)

        // if (targetNode === null || targetNode === undefined) {
        //     targetNode = 'none'
        // }

        // if (!$hardwareNodesLoaded) {
        //     fetchHardwareNodes()
        // }
    }

    function cancel() {
        // name = nameBefore
        // watts = wattsBefore
        data = structuredClone(dataBefore)
    }

    let configuredChanged = false
    let configuredData = structuredClone(data.singletonJson)


    function reactToOutput(modified: any) {
        if (!open) {
            return
        }

        console.dir(modified)
        // if (textarea.isEqualNode(document.activeElement) || preventReacttoOutput) {
        //     // TODO: is this even triggered?
        //     console.warn("Is active element, prevent cycle")
        //     return
        // }
        // textareaContent = `\n${JSON.stringify(data, null, 2)}`
        // lastOutput = data

        // console.dir(data)
        configuredData = structuredClone(modified)
        configuredChanged = true
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
        <!-- {#if $hardwareNodesLoaded} -->
        <!--     <Select bind:value={targetNode} label="Target Node"> -->
        <!--         {#each $hardwareNodes as node} -->
        <!--             {#if node === null} -->
        <!--                 <Option value={'none'}>None</Option> -->
        <!--             {:else} -->
        <!--                 <Option value={node.url}>{node.name}</Option> -->
        <!--             {/if} -->
        <!--         {/each} -->
        <!--     </Select> -->
        <!-- {:else} -->
        <!--     <Progress bind:loading={$loading} /> -->
        <!-- {/if} -->
        <!-- <br /> -->
        <!-- <br /> -->
        <Textfield bind:value={data.name} input$maxlength={30} label="Name" required>
            <svelte:fragment slot="helper">
                <CharacterCounter>0 / 30</CharacterCounter>
            </svelte:fragment>
        </Textfield>

        <!-- <Textfield bind:value={watts} label="Watts" type="number" /> -->

        <!-- Dynamic configurator -->
        <!-- bind:spec={driver.info.driver.info.config} -->
        <div>
            <DynamicConfigurator
                bind:spec={data.config.info.config}
                on:change={ (e) => reactToOutput(e.detail) }
                bind:inputData={data.singletonJson}
                topLevelLabel={`Device Configuration`}
            />

            <Button disabled={!configuredChanged} variant="outlined" on:click={saveDeviceConfig}>
                <Icon class="material-icons">save</Icon>
                <Label>Save</Label>
            </Button>
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
        <!-- <Button -->
        <!--     disabled={!nameDirty && !wattsDirty && !targetNodeDirty} -->
        <!--     use={[InitialFocus]} -->
        <!--     on:click={() => -->
        <!--         dispatch('modify', { -->
        <!--             name, -->
        <!--             watts, -->
        <!--             targetNode: targetNode === 'none' ? null : targetNode, -->
        <!--         })} -->
        <!-- > -->
        <!--     <Label>Modify</Label> -->
        <!-- </Button> -->
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
