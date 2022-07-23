<script lang="ts">
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        Header,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import IconButton from "@smui/icon-button";
    import { createEventDispatcher, onMount } from "svelte";
    import type {  ScheduleData } from "../main";
    import Inputs from "./Inputs.svelte";

    export let open = false;

    $: if (open) updatePrevious();

    // Event dispatcher
    const dispatch = createEventDispatcher();

    export let data: ScheduleData;
    let dataBefore: ScheduleData;

    function reset() {
        data = {
            name: dataBefore.name,
            hour: dataBefore.hour,
            minute: dataBefore.minute,
            homescriptCode: dataBefore.homescriptCode,
        };
        open = false;
    }

    function updatePrevious() {
        dataBefore = {
            name: data.name,
            hour: data.hour,
            minute: data.minute,
            homescriptCode: data.homescriptCode,
        };
    }

    onMount(updatePrevious);
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content" fullscreen>
    <Header>
        <Title id="title">Add Schedule</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
        <Inputs bind:data />
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
</style>
