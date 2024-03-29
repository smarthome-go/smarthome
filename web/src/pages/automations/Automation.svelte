<script lang="ts">
    import IconButton from '@smui/icon-button'
    import { createEventDispatcher, onMount } from 'svelte'
    import { createSnackbar, sleep } from '../../global'
    import AutomationInfo from './dialogs/AutomationInfo.svelte'
    import EditAutomation from './dialogs/EditAutomation.svelte'
    import {
        generateCronExpression,
        hmsLoaded,
        sunTimesLoaded,
        automationsLoaded,
        homescripts,
        loading,
        getTimeOfAutomation,
        triggerMetaData,
    } from './main'
    import type { automation, editAutomation, homescript } from './main'

    const days: string[] = ['su', 'mo', 'tu', 'we', 'th', 'fr', 'sa']

    // Event dispatcher
    const dispatch = createEventDispatcher()

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

    async function modifyAutomation(id: number, payload: editAutomation) {
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
            data.triggerCronExpression = generateCronExpression(
                payload.hour,
                payload.minute,
                payload.days,
            )
            // Updates the Homescriptdata of the automation
            const homescriptDataTemp = $homescripts.find(s => s.data.id === data.homescriptId)
            if (homescriptDataTemp !== undefined) homescriptData = homescriptDataTemp
        } catch (err) {
            $createSnackbar(`Could not modify automation: ${err}`)
        }
        $loading = false
    }

    let editOpen = false
    let infoOpen = false

    // Generates a 12h string from 24h time data
    let timeString = 'loading..'
    $: if ($automationsLoaded && $sunTimesLoaded)
        timeString =
            `${timeData.hours <= 12 ? timeData.hours : timeData.hours - 12}`.padStart(2, '0') +
            ':' +
            `${timeData.minutes}`.padStart(2, '0') +
            ` ${timeData.hours < 12 ? 'AM' : 'PM'}`

    // Update days and time
    $: if (
        $automationsLoaded &&
        $sunTimesLoaded &&
        (data.trigger === 'cron' || data.trigger === 'on_sunrise' || data.trigger === 'on_sunset')
    )
        timeData = getTimeOfAutomation(data)

    let triggerIntervalBuffer = 0
    let triggerIntervalUnit = undefined

    $: if (data.trigger === 'interval' && data.triggerInterval) {
        if (data.triggerInterval % (60 * 60 * 24) === 0) {
            triggerIntervalUnit = 'day'
            triggerIntervalBuffer = data.triggerInterval / (60 * 60 * 24)
        } else if (data.triggerInterval % (60 * 60) === 0) {
            triggerIntervalUnit = 'hour'
            triggerIntervalBuffer = data.triggerInterval / (60 * 60)
        } else if (data.triggerInterval % 60 === 0) {
            triggerIntervalUnit = 'minute'
            triggerIntervalBuffer = data.triggerInterval / 60
        } else {
            triggerIntervalUnit = 'second'
            triggerIntervalBuffer = data.triggerInterval
        }
    }

    async function handleEditAutomation(event) {
        const dataTemp = event.detail
        const dataTempEnabled = dataTemp.data.enabled
        const enabledStatusBefore = data.enabled
        const dataTempTimingMode = dataTemp.data.timingMode
        const triggerBefore = data.trigger

        // Modify the automation
        await modifyAutomation(dataTemp.id, dataTemp.data)
        data.disableOnce = dataTemp.data.disableOnce
        if (dataTempEnabled !== enabledStatusBefore || dataTempTimingMode !== triggerBefore) {
            dispatch('modify', null)
        }
    }

    onMount(async () => {
        while (!$hmsLoaded) await sleep(5)
        const homescriptDataTemp = $homescripts.find(s => s.data.id === data.homescriptId)
        if (homescriptDataTemp !== undefined && homescriptDataTemp !== null)
            homescriptData = homescriptDataTemp
    })
</script>

<EditAutomation
    bind:open={editOpen}
    {data}
    on:modify={handleEditAutomation}
    on:delete={() => dispatch('delete', null)}
/>

<AutomationInfo
    {triggerIntervalBuffer}
    {triggerIntervalUnit}
    {timeString}
    bind:data
    bind:open={infoOpen}
/>

<div class="automation mdc-elevation--z3" class:disabled={!data.enabled || data.disableOnce}>
    <!-- Top -->
    <div class="top">
        <span class="automation__name">
            {data.name}
            <i
                class="material-icons automation__indicator"
                class:disabled={!data.enabled}
                class:disabled-once={data.disableOnce && data.enabled}
            >
                {#if !data.enabled}
                    sync_disabled
                {:else if data.disableOnce}
                    sync_problem
                {:else}
                    published_with_changes
                {/if}
            </i>
        </span>
        <span class="automation__time">
            {#if data.trigger === 'cron'}
                At {timeString}
            {:else if data.trigger === 'on_sunrise' || data.trigger === 'on_sunset'}
                At {triggerMetaData[data.trigger].name.toLowerCase()}
                <div class="automation__time__mode">
                    <span class="text-hint">{timeString}</span>
                    <i class="material-icons trigger-icon">
                        {triggerMetaData[data.trigger].icon}
                    </i>
                </div>
            {:else if data.trigger === 'interval'}
                {triggerMetaData[data.trigger].name}
                <div class="automation__time__mode">
                    {#if triggerIntervalBuffer == 1}
                        <span class="text-hint">every {triggerIntervalUnit}</span>
                    {:else}
                        <span class="text-hint"
                            >every {triggerIntervalBuffer} {triggerIntervalUnit}s</span
                        >
                    {/if}
                    <i class="material-icons trigger-icon">
                        {triggerMetaData[data.trigger].icon}
                    </i>
                </div>
            {:else}
                <div class="automation__time__mode">
                    {triggerMetaData[data.trigger].name}
                    <i class="material-icons trigger-icon">
                        {triggerMetaData[data.trigger].icon}
                    </i>
                </div>
            {/if}
        </span>
        <!-- Days -->
        <span class="automation__days">
            {#if timeData.days.length === 7}
                <span class="day">every day <i class="material-icons">restart_alt</i> </span>
            {:else}
                {#each timeData.days.map(d => days[d]) as day}
                    <span class="day">{day}</span>
                {/each}
            {/if}
        </span>
    </div>

    <!-- Bottom -->
    <div class="bottom">
        <span class="automation__homescript text-hint">
            <span
                >{homescriptData.data.name}
                <!-- If the Homescripts are loaded, display the script's icon for nicer display -->
            </span>
            {#if hmsLoaded}
                <i class="material-icons automation__homescript__icon">
                    {homescriptData.data.mdIcon}
                </i>
            {/if}
        </span>
        <div class="bottom__buttons">
            <IconButton class="material-icons" on:click={() => (editOpen = true)}>edit</IconButton>
            <IconButton class="material-icons" on:click={() => (infoOpen = true)}>info</IconButton>
        </div>
    </div>
</div>

<style lang="scss">
    @use '../../mixins' as *;
    .automation {
        height: 9rem;

        // Was chosen because it looks best on 1080p
        width: 17.5rem;

        border-radius: 0.3rem;
        padding: 1rem;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        background-color: var(--clr-height-1-3);

        .trigger-icon {
            font-size: 1.2rem;
            color: var(--clr-text-hint);
        }

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
            font-size: 0.85rem;
            margin-bottom: 0.45rem;

            &__mode {
                display: flex;
                gap: 0.7rem;
                align-items: center;
                justify-content: center;

                span {
                    font-size: 0.8rem;
                }
            }
        }

        &__indicator {
            font-size: 1.3rem;
            color: var(--clr-success);
            opacity: 85%;

            &.disabled-once {
                color: var(--clr-warn);
                opacity: 100%;
                filter: brightness(110%);
            }

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
            margin-bottom: 0.2rem;
            font-weight: bold;
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
