<script lang="ts">
    import { Icon } from '@smui/button'
    import { onMount } from 'svelte'
    import { sleep } from '../../global'
    import {
    automation,
    hmsLoaded,
    homescript,
    homescripts,
    parseCronExpressionToTime
    } from './main'

    const days: string[] = ['su', 'mo', 'tu', 'we', 'th', 'fr', 'sa']

    export let data: automation

    let homescriptData: homescript = {
        owner: '',
        data: {
            code: '',
            description: '',
            id: '',
            mdIcon: '',
            name: '',
            quickActionsEnabled: false,
            schedulerEnabled: false,
        },
    }

    let timeData = {
        hours: 0,
        minutes: 0,
        days: [],
    }

    // Generates a 12h string from 24h time data
    let timeString = ''
    $: timeString =
        `${
            timeData.hours < 12 ? timeData.hours : timeData.hours - 12
        }`.padStart(2, '0') +
        ':' +
        `${timeData.minutes}`.padStart(2, '0') +
        ` ${timeData.hours < 12 ? 'AM' : 'PM'}`

    onMount(async () => {
        while (!$hmsLoaded) await sleep(5)
        const homescriptDataTemp = $homescripts.find(
            (s) => s.data.id === data.homescriptId
        )
        if (homescriptDataTemp !== undefined && homescriptDataTemp !== null)
            homescriptData = homescriptDataTemp
        // Set time
        timeData = parseCronExpressionToTime(data.cronExpression)
    })
</script>

<div class="automation mdc-elevation--z3">
    <!-- Top -->
    <div class="top">
        <span class="automation__name">{data.name}</span>
        <span class="automation__time">
            At {timeString}
            {#if timeData.days.length === 7}
                <span class="day"
                    >every day <i class="material-icons">restart_alt</i>
                </span>
            {/if}
        </span>
        <!-- Days -->
        <span class="automation__days">
            {#if timeData.days.length !== 7}
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
                <Icon class="material-icons">
                    {homescriptData.data.mdIcon}
                </Icon>
            {/if}
        </span>
    </div>
</div>

<style lang="scss">
    @use '../../mixins' as *;
    .automation {
        height: 9rem;
        width: 15rem;
        border-radius: 0.3rem;
        background-color: var(--clr-height-1-3);
        padding: 1rem;

        display: flex;
        flex-direction: column;
        justify-content: space-between;

        &__homescript {
            display: flex;
            gap: 0.5rem;
            font-size: 0.9rem;
        }

        &__time {
            display: flex;
            justify-content: space-between;
        }

        .top,
        .bottom {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
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
