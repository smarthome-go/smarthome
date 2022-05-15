<script lang="ts">
    import { Icon } from '@smui/button'
    import IconButton from '@smui/icon-button/src/IconButton.svelte'
    import { onMount } from 'svelte'
    import { createSnackbar,sleep } from '../../global'
    import EditAutomation from './dialogs/EditAutomation.svelte'
    import {
    addAutomation,
    automation,
    generateCronExpression,
    hmsLoaded,
    homescript,
    homescripts,
    loading,
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

    interface timeDataType {
        hours: number
        minutes: number
        days: number[]
    }

    let timeData: timeDataType = {
        hours: 0,
        minutes: 0,
        days: [],
    }

    async function modifyAutomation(id: number, payload: addAutomation) {
        $loading = true
        try {
            payload['id'] = id
            const res = await (
                await fetch('/api/automation/modify', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload),
                })
            ).json()
            if (!res.success) throw Error(res.error)
            data.cronExpression = generateCronExpression(
                payload.hour,
                payload.minute,
                payload.days
            )
            const homescriptDataTemp = $homescripts.find(
                (s) => s.data.id === data.homescriptId
            )
            if (homescriptDataTemp !== undefined)
                homescriptData = homescriptDataTemp
        } catch (err) {
            $createSnackbar(`Could not modify automation: ${err}`)
        }
        $loading = false
    }

    let editOpen = false

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
    })

    // Update days and time
    $: timeData = parseCronExpressionToTime(data.cronExpression)

    function handleEditAutomation(event) {
        const dataTemp = event.detail
        modifyAutomation(dataTemp.id, dataTemp.data).then()
    }
</script>

<EditAutomation bind:open={editOpen} {data} on:modify={handleEditAutomation} />

<div class="automation mdc-elevation--z3">
    <!-- Top -->
    <div class="top">
        <span class="automation__name">{data.name}</span>
        <span class="automation__time">
            At
            {timeString}
            <!-- {timeData.hours.toString().padStart(2, "0")}:{timeData.minutes.toString().padStart(2, "0")} -->
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
        <IconButton class="material-icons" on:click={() => (editOpen = true)}
            >edit</IconButton
        >
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

        .top {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }
        .bottom {
            display: flex;
            gap: 0.5rem;
            align-items: center;
            justify-content: space-between;
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
