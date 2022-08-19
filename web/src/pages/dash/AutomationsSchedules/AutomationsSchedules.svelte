<script lang="ts">
    import Box from "../Box.svelte";
    import { createSnackbar } from "../../../global";
    import { onMount } from "svelte";
    import type {
        automation,
        automationWrapper,
        Schedule as ScheduleType,
    } from "./types";
    import Schedule from "./Schedule.svelte";
    import Automation from "./Automation.svelte";
    import Button, { Label } from "@smui/button";

    let loading = false;

    let automations: automationWrapper[] = [];
    let automationsToday: automationWrapper[] = [];
    let automationsLoaded = false;

    const now = new Date();

    // Fetches the current automations from the server
    async function loadAutomations() {
        loading = true;
        try {
            const res = await (
                await fetch("/api/automation/list/personal")
            ).json();

            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            (res as automation[]).forEach((a) => {
                const timeData = parseCronExpressionToTime(a.cronExpression);
                automations = [
                    ...automations,
                    {
                        data: a,
                        hours: timeData.hours,
                        minutes: timeData.minutes,
                        days: timeData.days,
                    },
                ];
            });
            automationsToday = automations.filter((a) => {
                // Filter out any disabled automations
                if (!a.data.enabled) return false

                // Filter out any automations from not today
                if (!a.days.includes(now.getDay())) return false;

                // Only display the automations which are still coming
                return (
                    a.hours > now.getHours() ||
                    (a.hours === now.getHours() &&
                        a.minutes >= now.getMinutes())
                );
            });
            automationsLoaded = true;
        } catch (err) {
            $createSnackbar(`Could not load automations: ${err}`);
        }
        loading = false;
    }

    // Parses a valid cron-expression, if it is invalid, an error is thrown
    export function parseCronExpressionToTime(expr: string): {
        hours: number;
        minutes: number;
        days: number[];
    } {
        if (expr === "* * * * *") return { days: [], hours: 0, minutes: 0 };
        const rawExpr = expr.split(" ");
        if (rawExpr.length != 5)
            throw Error(`Invalid cron-expression: '${expr}'`);
        // Days
        let days: number[] = [];
        if (rawExpr[4] === "*") days = [0, 1, 2, 3, 4, 5, 6];
        else days = rawExpr[4].split(",").map((d) => parseInt(d));
        return {
            hours: parseInt(rawExpr[1]),
            minutes: parseInt(rawExpr[0]),
            days: days,
        };
    }

    let schedules: ScheduleType[] = [];
    let schedulesLoaded = false;

    // Fetches the current schedules from the server
    async function loadSchedules() {
        loading = true;
        try {
            const res = await (
                await fetch("/api/scheduler/list/personal")
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            schedules = res
            schedulesLoaded = true;
        } catch (err) {
            $createSnackbar(`Could not load schedules: ${err}`);
        }
        loading = false;
    }

    onMount(async () => {
        await loadAutomations();
        await loadSchedules();
    });
</script>

<Box bind:loading>
    <span slot="header">Schedules and Automations</span>
    <div class="content" slot="content">
        <div class="content__automations">
            {#if automationsLoaded && automationsToday.length === 0}
                <div class="content__automations__empty">
                    <span class="content__automations__empty__title">
                        No Automations
                    </span>
                    <span class="text-hint"> No automations running Today </span>
                    <Button variant="outlined" href='/automations'>
                        <Label>Create</Label>
                    </Button>
                </div>
            {:else}
                <span class="content__automations__title">
                    {automationsToday.length} Automation{automationsToday.length !==
                    1
                        ? "s"
                        : ""}
                    (Upcoming)
                </span>
                <div class="content__automations__list">
                    {#each automationsToday as data}
                        <Automation bind:data />
                    {/each}
                </div>
            {/if}
        </div>
        <div class="content__schedules">
            {#if schedulesLoaded && schedules.length === 0}
                <div class="content__schedules__empty">
                    <span class="content__schedules__empty__title">
                        No Schedules
                    </span>
                    <span class="text-hint">Nothing planned soon</span>
                    <Button variant="outlined" href='/scheduler'>
                        <Label>Plan</Label>
                    </Button>
                </div>
            {:else}
                <span class="content__schedules__title">
                    {schedules.length} Schedule{schedules.length !== 1
                        ? "s"
                        : ""} (Planned)
                </span>
                <div class="content__schedules__list">
                    {#each schedules as data}
                        <Schedule bind:data />
                    {/each}
                </div>
            {/if}
        </div>
    </div>
</Box>

<style lang="scss">
    .content {
        display: flex;
        gap: 1rem;

        &__automations {
            width: 50%;

            &__empty {
                margin-top: 0.8rem;
                display: flex;
                flex-direction: column;
                align-items: flex-start;

                &__title {
                    font-weight: bold;
                }

                .text-hint {
                    font-size: 0.9rem;
                    margin-bottom: 0.8rem;
                }
            }

            &__title {
                color: var(--clr-text-hint);
                font-size: 0.9rem;
            }

            &__list {
                margin-top: 0.5rem;
                display: flex;
                flex-direction: column;
                gap: 0.5rem;
                height: 2rem;
            }
        }

        &__schedules {
            width: 50%;

            &__empty {
                margin-top: 0.8rem;
                display: flex;
                flex-direction: column;
                align-items: flex-start;

                &__title {
                    font-weight: bold;
                }

                .text-hint {
                    font-size: 0.9rem;
                    margin-bottom: 0.8rem;
                }
            }

            &__title {
                color: var(--clr-text-hint);
                font-size: 0.9rem;
            }

            &__list {
                margin-top: 0.5rem;
                display: flex;
                flex-direction: column;
                gap: 0.5rem;
            }
        }
    }
</style>
