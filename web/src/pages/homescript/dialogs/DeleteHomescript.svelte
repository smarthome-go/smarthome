<script lang="ts">
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import { createEventDispatcher } from "svelte";
    import type { homescriptData } from "../../../homescript";

    export let open = false;
    export let data: homescriptData;

    // Event dispatcher
    const dispatch = createEventDispatcher();
</script>

<!-- Deletion confirmation dialog -->
<Dialog
    bind:open
    aria-labelledby="confirmation-title"
    aria-describedby="confirmation-content"
    slot="over"
>
    <Title id="confirmation-title">Confirm Deletion of '{data.name}'</Title>
    <Content id="confirmation-content">
        Deletion may cause unintended consequences, these include:
        <ul class="consequences">
            <li>
                Breaking Homescripts which use <code>exec('{data.id}')</code>
            </li>
            <li>
                Removing all automations which have '{data.id}' as their
                target
            </li>
        </ul>
        Please only proceed if you are able to identify all consequences of your
        action. Are you sure that you want to proceed?
    </Content
    >
    <Actions>
        <Button
            on:click={() => {
                dispatch("delete", { id: data.id });
            }}
        >
            <Label>Delete</Label>
        </Button>

        <Button defaultAction use={[InitialFocus]}>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    .consequences {
        margin: 0.55rem 0;

        code {
            border-radius: 0.3rem;
            padding: 0.05rem 0.2rem;
            background-color: var(--clr-height-0-2);
            color: var(--clr-primary);
            font-size: 0.9rem;
            font-family: "Jetbrains Mono", monospace;
        }
    }
</style>
