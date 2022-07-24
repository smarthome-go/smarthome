<script lang="ts">
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import type { automation } from "../main";

    export let open = false;

    export let data: automation;

    const timingModes = {
        normal: { name: "Time set manually", icon: "schedule" },
        sunrise: { name: "Local sunrise", icon: "wb_twilight" },
        sunset: { name: "Local sunset", icon: "nights_stay" },
    };
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
                <!-- Cron Description -->
                <h6>Timing</h6>

                <div class="timing-mode">
                    {timingModes[data.timingMode].name}
                    <i class="material-icons">
                        {timingModes[data.timingMode].icon}
                    </i>
                </div>

                {data.cronDescription}
                <br />
                <span class="text-disabled">
                    Cron-expression: <code>{data.cronExpression}</code>
                </span>
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
        font-family: "Jetbrains Mono", monospace;
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
    .timing-mode {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }
</style>
