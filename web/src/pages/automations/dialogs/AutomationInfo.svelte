<script lang="ts">
    import Button, { Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import type { automation } from '../main'
    import { triggerMetaData } from '../main'

    export let open = false

    export let data: automation
    export let timeString: string = undefined
    export let triggerIntervalBuffer: number = null
    export let triggerIntervalUnit: string = null
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">{data.name}</Title>
    <Content id="content">
        <div class="container">
            <!-- Do not show an empty description -->
            {#if data.description.length > 0}
                <div>
                    <!-- Description -->
                    <h6>Description</h6>
                    {data.description}
                </div>
            {/if}
            <div>
                <h6>Details</h6>
                Internal ID:
                {data.id}
            </div>
            <div>
                <!-- Trigger Information -->
                <h6>Trigger</h6>

                <div class="trigger-mode">
                    {triggerMetaData[data.trigger].name}
                    <i class="material-icons">
                        {triggerMetaData[data.trigger].icon}
                    </i>
                </div>

                {#if data.trigger == 'cron'}
                    {data.cronDescription}
                    <br />
                    <span class="text-disabled">
                        Cron-expression: <code>{data.triggerCronExpression}</code>
                    </span>
                {:else if data.trigger === 'on_sunrise' || data.trigger === 'on_sunset'}
                    At {timeString}
                {:else if data.trigger === 'interval'}
                    {#if triggerIntervalBuffer == 1}
                        <span class="text-hint">every {triggerIntervalUnit}</span>
                    {:else}
                        <span class="text-hint"
                            >every {triggerIntervalBuffer} {triggerIntervalUnit}s</span
                        >
                    {/if}
                {:else if data.trigger === 'on_login' || data.trigger === 'on_logout' || data.trigger === 'on_shutdown' || data.trigger === 'on_notification'}
                    <!-- Ignore these -->
                {:else}
                    Trigger {data.trigger} not supported
                {/if}
            </div>
        </div>
    </Content>
    <Actions>
        <Button use={[InitialFocus]}>
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    h6 {
        margin: 0.3rem 0;
    }
    code {
        font-family: 'Jetbrains Mono', monospace;
        font-size: 0.9rem;
    }
    .container {
        display: flex;
        flex-direction: column;
        gap: 1.45rem;

        h6 {
            font-size: 1rem;
            text-decoration: underline;
            margin: 0;
        }
    }
    .trigger-mode {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }
</style>
