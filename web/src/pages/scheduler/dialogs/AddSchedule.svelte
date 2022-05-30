<script lang="ts">
    import Button, { Icon, Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        Header,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import IconButton from "@smui/icon-button";
    import { hasPermission } from "../../../global";
    import { createEventDispatcher, onMount } from "svelte";
    import { schedules } from "../main";
    import type { ScheduleData, addSchedule } from "../main";
    import Inputs from "./Inputs.svelte";

    export let open = false;
    let hasHomescriptPermission = false;

    // Event dispatcher
    const dispatch = createEventDispatcher();

    // Binded to the `Inputs.svelte` component
    let data: addSchedule = {
        hour: 0,
        minute: 0,
        name: "",
        homescriptCode: "",
    };

    function reset() {
        data = {
            hour: 0,
            minute: 0,
            name: "",
            homescriptCode: "",
        };
        open = false;
    }

    onMount(async () => {
        hasHomescriptPermission = await hasPermission("homescript");
    });
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content" fullscreen>
    <Header>
        <Title id="title">Add Schedule</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
        {#if hasHomescriptPermission}
            <p>
                You are missing the <code>homescript</code> permission.
                <br />
                This permission is required in order to use the scheduler.
                <span class="text-hint"
                    >You can also use the CLI to create Homescripts. <a
                        href="https://github.com/smarthome-go/cli"
                        target="_blank">learn more</a
                    ></span
                >
            </p>
        {:else}
            <Inputs bind:data />
        {/if}
    </Content>
    <Actions>
        <Button on:click={reset}>
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={data.name == "" || data.homescriptCode.length == 0}
            use={[InitialFocus]}
            on:click={() => {
                dispatch("add", data);
                // Reset values after creation
                reset();
            }}
        >
            <Label>Create</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    .text-hint {
        font-size: 0.9rem;
        display: block;
    }
    a {
        color: var(--clr-primary);
        opacity: 90%;
    }
</style>
