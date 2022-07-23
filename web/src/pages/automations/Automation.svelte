<script lang="ts">
    import IconButton from "@smui/icon-button/src/IconButton.svelte";
    import { createEventDispatcher, onMount } from "svelte";
    import { createSnackbar, sleep } from "../../global";
    import AutomationInfo from "./dialogs/AutomationInfo.svelte";
    import EditAutomation from "./dialogs/EditAutomation.svelte";
    import {
        addAutomation,
        automation,
        generateCronExpression,
        hmsLoaded,
        homescript,
        homescripts,
        loading,
        parseCronExpressionToTime,
    } from "./main";

    // Event dispatcher
    const dispatch = createEventDispatcher();

    const days: string[] = ["su", "mo", "tu", "we", "th", "fr", "sa"];
    const timingModes = {
        normal: { name: "", icon: "schedule" },
        sunrise: { name: "Sunrise", icon: "wb_twilight" },
        sunset: { name: "Sunset", icon: "nights_stay" },
    };

    export let data: automation;

    let homescriptData: homescript = {
        owner: "",
        data: {
            code: "",
            description: "",
            id: "",
            mdIcon: "",
            name: "",
            quickActionsEnabled: false,
            schedulerEnabled: false,
        },
    };

    interface timeDataType {
        hours: number;
        minutes: number;
        days: number[];
    }

    let timeData: timeDataType = {
        hours: 0,
        minutes: 0,
        days: [],
    };

    async function modifyAutomation(id: number, payload: addAutomation) {
        $loading = true;
        try {
            payload["id"] = id;
            const res = await (
                await fetch("/api/automation/modify", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(payload),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            data.cronExpression = generateCronExpression(
                payload.hour,
                payload.minute,
                payload.days
            );
            const homescriptDataTemp = $homescripts.find(
                (s) => s.data.id === data.homescriptId
            );
            if (homescriptDataTemp !== undefined)
                homescriptData = homescriptDataTemp;
        } catch (err) {
            $createSnackbar(`Could not modify automation: ${err}`);
        }
        $loading = false;
    }

    let editOpen = false;
    let infoOpen = false;

    // Generates a 12h string from 24h time data
    let timeString = "";
    $: timeString =
        `${
            timeData.hours <= 12 ? timeData.hours : timeData.hours - 12
        }`.padStart(2, "0") +
        ":" +
        `${timeData.minutes}`.padStart(2, "0") +
        ` ${timeData.hours < 12 ? "AM" : "PM"}`;

    onMount(async () => {
        while (!$hmsLoaded) await sleep(5);
        const homescriptDataTemp = $homescripts.find(
            (s) => s.data.id === data.homescriptId
        );
        if (homescriptDataTemp !== undefined && homescriptDataTemp !== null)
            homescriptData = homescriptDataTemp;
    });

    // Update days and time
    $: timeData = parseCronExpressionToTime(data.cronExpression);

    async function handleEditAutomation(event) {
        const dataTemp = event.detail;
        const dataTempEnabled = dataTemp.data.enabled;
        const enabledStatusBefore = data.enabled;
        const dataTempTimingMode = dataTemp.data.timingMode;
        const timingModeBefore = data.timingMode;
        await modifyAutomation(dataTemp.id, dataTemp.data);
        if (
            dataTempEnabled !== enabledStatusBefore ||
            dataTempTimingMode !== timingModeBefore
        )
            dispatch("modify", null);
    }

    function handleDeleteAutomation(_event) {
        dispatch("delete", null);
    }
</script>

<EditAutomation
    bind:open={editOpen}
    {data}
    on:modify={handleEditAutomation}
    on:delete={handleDeleteAutomation}
/>

<AutomationInfo bind:data bind:open={infoOpen} />

<div class="automation mdc-elevation--z3" class:disabled={!data.enabled}>
    <!-- Top -->
    <div class="top">
        <span class="automation__name">
            {data.name}
            <i
                class="material-icons automation__indicator"
                class:disabled={!data.enabled}
            >
                {#if data.enabled}
                    published_with_changes
                {:else}
                    sync_disabled
                {/if}
            </i>
        </span>
        <span class="automation__time">
            At
            {#if data.timingMode === "normal"}
                {timeString}
            {:else}
                {timingModes[data.timingMode].name}
                <div class="automation__time__mode">
                    <span class="text-hint">{timeString}</span>
                    <i class="material-icons">
                        {timingModes[data.timingMode].icon}
                    </i>
                </div>
            {/if}
        </span>
        <!-- Days -->
        <span class="automation__days">
            {#if timeData.days.length === 7}
                <span class="day"
                    >every day <i class="material-icons">restart_alt</i>
                </span>
            {:else}
                {#each timeData.days.map((d) => days[d]) as day}
                    <span class="day">{day}</span>
                {/each}
            {/if}
        </span>
    </div>

    <!-- Bottom -->
    <div class="bottom">
        <span class="automation__homescript text-hint">
            <span
                >Script:
                {homescriptData.data.name}
                <!-- If Homescript is loaded, display the script's icon for better readability -->
            </span>
            {#if hmsLoaded}
                <i class="material-icons automation__homescript__icon">
                    {homescriptData.data.mdIcon}
                </i>
            {/if}
        </span>
        <div class="bottom__buttons">
            <IconButton
                class="material-icons"
                on:click={() => (editOpen = true)}>edit</IconButton
            >
            <IconButton
                class="material-icons"
                on:click={() => (infoOpen = true)}>info</IconButton
            >
        </div>
    </div>
</div>

<style lang="scss">
    @use "../../mixins" as *;
    .automation {
        height: 9rem;
        width: 15rem;
        border-radius: 0.3rem;
        padding: 1rem;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        background-color: var(--clr-height-1-3);

        &.disabled {
            opacity: 75%;
        }

        &__homescript {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-size: 0.9rem;

            &__icon {
                font-size: 1.2rem;
            }
        }

        &__time {
            display: flex;
            justify-content: space-between;
            font-size: .85rem;
            margin-bottom: .45rem;

            &__mode {
                display: flex;
                gap: 0.6rem;
                align-items: center;

                span {
                    font-size: 0.8rem;
                }
            }
        }

        &__indicator {
            font-size: 1.3rem;
            color: var(--clr-success);
            opacity: 85%;

            &.disabled {
                color: var(--clr-error);
                opacity: 100%;
                filter: brightness(110%);
            }
        }

        &__name {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: .2rem;
        }

        .top {
            display: flex;
            flex-direction: column;
            //   gap: 0.5rem;
        }
        .bottom {
            display: flex;
            gap: 0.5rem;
            align-items: center;
            justify-content: space-between;

            &__buttons {
                display: flex;
            }
        }

        &__days {
            display: flex;
            gap: 0.2rem;
        }

        .day {
            border-radius: 0.6rem;
            background-color: var(--clr-height-3-4);
            color: var(--clr-primary);
            opacity: 70%;
            padding: 0 0.5rem;
            font-size: 0.8rem;
            cursor: default;

            display: flex;
            align-items: center;
            gap: 0.4rem;

            i {
                font-size: 1rem;
            }
        }

        @include mobile {
            width: 80vw;
        }
    }
</style>
