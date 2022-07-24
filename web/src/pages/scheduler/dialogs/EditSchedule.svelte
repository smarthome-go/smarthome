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
    import { createSnackbar } from "../../../global";
    import { onMount } from "svelte";
    import { loading, Schedule, ScheduleData } from "../main";
    import Inputs from "./Inputs.svelte";

    export let open = false;

    $: if (open) upDatePrevious();

    export let data: Schedule;
    let dataBefore: ScheduleData;

    function reset() {
        data.data = {
            name: dataBefore.name,
            hour: dataBefore.hour,
            minute: dataBefore.minute,
            homescriptCode: dataBefore.homescriptCode,
        };
    }

    function upDatePrevious() {
        dataBefore = {
            name: data.data.name,
            hour: data.data.hour,
            minute: data.data.minute,
            homescriptCode: data.data.homescriptCode,
        };
    }

    // Modifies the data of the current schedule
    async function modifySchedule() {
        $loading = true;
        try {
            const res = await (
                await fetch("/api/scheduler/modify", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        id: data.id,
                        data: data.data,
                    }),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            upDatePrevious();
        } catch (err) {
            $createSnackbar(`Could not modify schedule: ${err}`);
            reset();
        }
        $loading = false;
    }

    onMount(upDatePrevious);
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content" fullscreen>
    <Header>
        <Title id="title">Add Schedule</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
        <Inputs bind:data={data.data} />
    </Content>
    <Actions>
        <Button on:click={reset} use={[InitialFocus]}>
            <Label>Cancel</Label>
        </Button>
        <Button
            disabled={data.data.name == "" ||
                data.data.homescriptCode.length == 0 ||
                JSON.stringify(data.data) === JSON.stringify(dataBefore)}
            on:click={modifySchedule}
        >
            <Label>Update</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
</style>
