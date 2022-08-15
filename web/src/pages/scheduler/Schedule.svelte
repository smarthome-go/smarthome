<script lang="ts">
    import IconButton from "@smui/icon-button/src/IconButton.svelte";
    import { lintHomescriptCode } from "../../homescript";
    import { createEventDispatcher, onMount } from "svelte";
    import ConfirmDeletion from "./dialogs/ConfirmDeletion.svelte";
    import EditSchedule from "./dialogs/EditSchedule.svelte";
    import {
        Schedule,
        timeUntilExecutionText,
        switches,
        homescripts,
    } from "./main";
    import Progress from "../../components/Progress.svelte";

    export let data: Schedule;

    // Specifies whether the edit dialog should be open or not
    let editOpen = false;
    // Specifies whether the delete dialog should be open or not
    let deleteOpen = false;

    // Event dispatcher
    const dispatch = createEventDispatcher();

    // Generates a 12h string from 24h time data
    let timeString = "";
    $: timeString =
        `${
            data.data.hour <= 12 ? data.data.hour : data.data.hour - 12
        }`.padStart(2, "0") +
        ":" +
        `${data.data.minute}`.padStart(2, "0") +
        ` ${data.data.hour < 12 ? "AM" : "PM"}`;

    let timeRunning = false;
    let timeUntilString = "";

    // Recursive function which updates the `timeUntilString` every 1000ms
    function updateTimeUntilExecutionText() {
        timeUntilString = timeUntilExecutionText(
            new Date(),
            data.data.hour,
            data.data.minute
        );
        timeRunning =
            data.data.hour === new Date().getHours() &&
            data.data.minute === new Date().getMinutes();

        // If the schedule is assumed to be executing, hide it after 5 seconds
        if (timeRunning && !editOpen) {
            setTimeout(() => dispatch("hide"), 5000);
        }

        setTimeout(updateTimeUntilExecutionText, 1000);
    }

    // Start the time updater
    onMount(updateTimeUntilExecutionText);
</script>

<EditSchedule bind:data bind:open={editOpen} />
<ConfirmDeletion
    bind:open={deleteOpen}
    name={data.data.name}
    on:confirm={() => {
        dispatch("delete");
    }}
/>

<div class="schedule">
    <div class="top">
        <span class="schedule__name">{data.data.name}</span>
        <div class="schedule__time">
            <span class="schedule__time__at">
                At {timeString}
            </span>
            <span class="schedule__time__in">
                {#if timeRunning}
                    {#if editOpen}
                        Now
                    {:else}
                        Executing
                    {/if}
                {:else}
                    {timeUntilString}
                {/if}
            </span>
        </div>
    </div>

    <div class="bottom">
        <span class="schedule__target-indicator">
            {#if data.data.targetMode === "hms"}
                Target: Homescript
                <i class="material-icons">list</i>
            {:else if data.data.targetMode === "code"}
                Target: Code
                <i class="material-icons">code</i>
            {:else if data.data.targetMode === "switches"}
                Target: Switches
                <i class="material-icons">power</i>
            {/if}
        </span>

        <div class="schedule__target">
            {#if data.data.targetMode === "hms"}
                <div class="schedule__target__hms">
                    <span class="schedule__target__hms__name"
                        >{$homescripts.find(
                            (h) => h.data.id === data.data.homescriptTargetId
                        ).data.name}
                        <i class="material-icons">
                            {$homescripts.find(
                                (h) =>
                                    h.data.id === data.data.homescriptTargetId
                            ).data.mdIcon}
                        </i>
                    </span>

                    <span class="schedule__target__hms__lint">
                        {#await lintHomescriptCode($homescripts.find((h) => h.data.id === data.data.homescriptTargetId).data.code, [])}
                            <Progress type="circular" loading={true} />
                        {:then res}
                            {res.success ? "Working" : res.error[0].errorType}
                            <i
                                class="material-icons"
                                class:passing={res.success}
                                >{res.success ? "check" : "cancel"}</i
                            >
                        {/await}
                    </span>
                </div>
            {:else if data.data.targetMode === "code"}
                <div class="schedule__target__code">
                    <span>
                        Lines of code:
                        <code>
                            {data.data.homescriptCode.split("\n").length}
                        </code>
                    </span>
                    <span class="schedule__target__code__indicator">
                        {#await lintHomescriptCode(data.data.homescriptCode, [])}
                            <Progress type="circular" loading={true} />
                        {:then res}
                            {res.success ? "Working" : res.error[0].errorType}
                            <i
                                class="material-icons"
                                class:passing={res.success}
                                >{res.success ? "check" : "cancel"}</i
                            >
                        {/await}
                    </span>
                </div>
            {:else if data.data.targetMode === "switches"}
                <div class="schedule__target__switches">
                    {#each data.data.switchJobs as sw (sw.switchId)}
                        <div
                            class="schedule__target__switches__switch"
                            class:on={sw.powerOn}
                        >
                            <span>
                                {$switches.find((s) => s.id === sw.switchId)
                                    .name}
                            </span>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
        <div class="schedule__buttons">
            <IconButton
                class="material-icons"
                on:click={() => (deleteOpen = true)}>cancel</IconButton
            >
            <IconButton
                class="material-icons"
                on:click={() => (editOpen = true)}>edit</IconButton
            >
            <IconButton class="material-icons">info</IconButton>
        </div>
    </div>
</div>

<style lang="scss">
    .schedule {
        height: 12rem;

        // Was chosen because it looks best on 1080p
        width: 17.5rem;

        border-radius: 0.3rem;
        padding: 1rem;
        display: flex;
        flex-direction: column;
        justify-content: space-between;

        background-color: var(--clr-height-1-3);

        &__name {
            font-weight: bold;
        }

        &__time {
            font-size: 0.85rem;

            span {
                display: block;
            }

            &__in {
                font-size: 0.85rem;
                color: var(--clr-text-hint);
            }
        }

        &__target-indicator {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-size: 0.85rem;
            color: var(--clr-text-hint);

            i {
                font-size: 1.3rem;
            }
        }

        &__target {
            margin-bottom: 0.1rem;

            &__hms {
                display: flex;
                justify-content: space-between;

                &__name {
                    display: flex;
                    align-items: center;
                    gap: 0.5rem;

                    i {
                        font-size: 1.2rem;
                    }
                }

                &__lint {
                    display: flex;
                    align-items: center;
                    gap: 0.3rem;
                    color: var(--clr-text-hint);
                    font-size: 0.9rem;

                    i {
                        color: var(--clr-error);
                        font-size: 1.2rem;

                        &.passing {
                            color: var(--clr-success);
                        }
                    }
                }
            }

            &__code {
                display: flex;
                justify-content: space-between;

                &__indicator {
                    display: flex;
                    align-items: center;
                    gap: 0.3rem;
                    color: var(--clr-text-hint);
                    font-size: 0.9rem;

                    i {
                        color: var(--clr-error);
                        font-size: 1.2rem;

                        &.passing {
                            color: var(--clr-success);
                        }
                    }
                }
            }

            &__switches {
                display: flex;
                gap: 0.2rem;
                flex-wrap: nowrap;
                overflow-x: hidden;

                &__switch {
                    border-radius: 0.6rem;
                    background-color: var(--clr-height-3-4);
                    opacity: 70%;
                    padding: 0 0.5rem;
                    font-size: 0.8rem;
                    cursor: default;
                    display: flex;
                    align-items: center;
                    gap: 0.4rem;
                    max-width: 5rem;
                    color: var(--clr-error);

                    &.on {
                        color: var(--clr-success);
                    }

                    span {
                        overflow: hidden;
                        white-space: nowrap;
                        text-overflow: ellipsis;
                    }
                }
            }
        }

        &__buttons {
            margin-top: 1rem;
        }
    }
</style>
