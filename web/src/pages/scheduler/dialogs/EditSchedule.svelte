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

    export let data: Schedule = {
        id: 0,
        owner: "",
        data: {
            name: "",
            hour: 0,
            minute: 0,
            targetMode: "hms",
            homescriptCode: "",
            homescriptTargetId: "",
            switchJobs: [],
        },
    };

    let dataBefore: ScheduleData = {
        name: "",
        hour: 0,
        minute: 0,
        targetMode: "hms",
        homescriptCode: "",
        homescriptTargetId: "",
        switchJobs: [],
    };

    function reset() {
        data.data = {
            name: dataBefore.name,
            hour: dataBefore.hour,
            minute: dataBefore.minute,
            targetMode: dataBefore.targetMode,
            homescriptCode: dataBefore.homescriptCode,
            homescriptTargetId: dataBefore.homescriptTargetId,
            switchJobs: dataBefore.switchJobs,
        };
    }

    function upDatePrevious() {
        dataBefore = {
            name: data.data.name,
            hour: data.data.hour,
            minute: data.data.minute,
            targetMode: data.data.targetMode,
            homescriptCode: data.data.homescriptCode,
            homescriptTargetId: data.data.homescriptTargetId,
            switchJobs: data.data.switchJobs,
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
                JSON.stringify(data.data) === JSON.stringify(dataBefore)}
            on:click={modifySchedule}
        >
            <Label>Update</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
</style>
