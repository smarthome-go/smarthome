<script lang="ts">
    import Dialog, {
        Actions,
        Content,
        Header,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import { createEventDispatcher } from "svelte";
    import type { homescriptData } from "../main";
    import Inputs from "./Inputs.svelte";

    export let open = false;

    // Event dispatcher
    const dispatch = createEventDispatcher();

    // Bound to the `Inputs.svelte` component
    let data: homescriptData = {
        id: "",
        name: "",
        description: "",
        mdIcon: "",
        code: "",
        quickActionsEnabled: false,
        schedulerEnabled: false,
    };

    function reset() {
        data = {
            id: "",
            name: "",
            description: "",
            mdIcon: "",
            code: "",
            quickActionsEnabled: false,
            schedulerEnabled: false,
        };
    }

    function submit() {
        dispatch("add", { data });
        // Reset data after creation
        reset();
        open = false;
    }
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content" fullscreen>
    <Header>
        <Title id="title">Add Homescript</Title>
    </Header>
    <Content id="content">
        <Inputs bind:data />
    </Content>
    <Actions>
        <Button
            on:click={() => {
                reset();
                open = false;
            }}
        >
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={true}
            use={[InitialFocus]}
            on:click={() => {
                submit();
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
</style>
