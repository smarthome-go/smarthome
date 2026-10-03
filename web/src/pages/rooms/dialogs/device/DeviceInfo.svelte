<script lang="ts">
    import Button, { Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import type { HydratedDeviceResponse } from '../../../../device';

    export let open = false

    export let data: HydratedDeviceResponse = null
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Device Information</Title>
    <Content id="content">
        <ul>
            <li>
                ID: <code>{data.shallow.id}</code>
            </li>
            <li>
                Type: <code>{data.shallow.type}</code>
            </li>
            <li>
                Name: <code>{data.shallow.name}</code>
            </li>
            <li>
                ModelID: <code>{data.shallow.modelId}</code>
            </li>
            <li>
                VendorID: <code>{data.shallow.vendorId}</code>
            </li>
            <li>
                RoomID: <code>{data.shallow.roomId}</code>
            </li>

            {#if data.extractions.dimmables !== null}
                <li>
                    Dimmables: <code>[{data.extractions.dimmables.map(d => `${d.label}: ${d.range}: ${d.value}`).join(", ")}]</code>
                </li>
            {/if}

            {#if data.extractions.powerInformation != null}
                <li>
                    Power: <code>PowerOn: {data.extractions.powerInformation.state}: {data.extractions.powerInformation.powerDrawWatts} Watts</code>
                </li>
            {/if}
        </ul>
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
