<script lang="ts">
    import Button, { Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import type { FetchedDriver } from '../driver';

    export let open = false
    export let data: FetchedDriver = null
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Device Information</Title>
    <Content id="content">
        <ul>
            <li>
                ID: <code>{data.driver.vendorId}:{data.driver.modelId}</code>
            </li>
            <li>
                Version: <code>{data.driver.version}</code>
            </li>
            <li>
                Name: {data.driver.name}
            </li>

            {#if data.info.driver.capabilities !== null}
                <li>
                    Capabilities: <code>{data.info.driver.capabilities.join(', ')}</code>
                </li>
            {/if}

            {#if data.validationErrors.length !== 0}
                <li>
                    Errors: <code>{data.validationErrors.map(e => e.message).join('; ')}</code>
                </li>
            {/if}
        </ul>
    </Content>
    <Actions>
        <Button defaultAction use={[InitialFocus]}>
            <Label>Done</Label>
        </Button>
    </Actions>
</Dialog>

<style style="scss">
    li {
        padding: .2rem 0;
    }

    code {
        background-color: var(--clr-height-0-3);
        padding: 0.1rem 0.5rem;
        border-radius: 0.3rem;
    }
</style>
