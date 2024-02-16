<script lang="ts">
    import Button, { Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import Progress from '../../../../../src/components/Progress.svelte'
    // import { fetchHardwareNodes, loading, hardwareNodesLoaded, hardwareNodes } from './main'
    import type { DeviceResponse } from '../../main';

    export let open = false

    export let data: DeviceResponse = null

    // export function show() {
    //     open = true

        // if (!$hardwareNodesLoaded) {
        //     fetchHardwareNodes().then(() => {
        //         setNodeLabel()
        //     })
        // } else {
        //     setNodeLabel()

    // function setNodeLabel() {
    //     if (targetNode === null || targetNode === undefined) {
    //         targetNodeLabel = 'None'
    //     } else {
    //         targetNodeLabel = $hardwareNodes.find(h => {
    //             if (h === undefined || h === null) {
    //                 return false
    //             }
    //             return h.url === targetNode
    //         }).name
    //     }
    // }
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Device Information</Title>
    <Content id="content">
        <ul>
            <li>
                ID: <code>{data.id}</code>
            </li>
            <li>
                Type: <code>{data.type}</code>
            </li>
            <li>
                Name: <code>{data.name}</code>
            </li>
            <li>
                ModelID: <code>{data.modelId}</code>
            </li>
            <li>
                VendorID: <code>{data.vendorId}</code>
            </li>
            <li>
                RoomID: <code>{data.roomId}</code>
            </li>

            {#if data.dimmables !== null}
                <li>
                    Dimmables: <code>[{data.dimmables.map(d => `${d.label}: ${d.range}: ${d.value}`).join(", ")}]</code>
                </li>
            {/if}

            {#if data.powerInformation != null}
                <li>
                    Power: <code>PowerOn: {data.powerInformation.state}: {data.powerInformation.powerDrawWatts} Watts</code>
                </li>
            {/if}
        </ul>
        <!-- Name: {name} -->
        <!-- <br /> -->
        <!-- Watts: {watts} -->
        <!-- <br /> -->
        <!-- {#if $hardwareNodesLoaded} -->
        <!--     Target Node: {targetNodeLabel} -->
        <!-- {:else} -->
        <!--     <Progress bind:loading={$loading} /> -->
        <!-- {/if} -->
        <br />
    </Content>
    <Actions>
        <Button defaultAction use={[InitialFocus]}>
            <Label>Done</Label>
        </Button>
    </Actions>
</Dialog>

<style style="scss">
    code {
        background-color: var(--clr-height-0-3);
        padding: 0.1rem 0.5rem;
        border-radius: 0.3rem;
    }
</style>
