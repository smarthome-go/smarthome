<!-- Contains the Input elements used by `AddAutomation` and `EditAutomation` -->
<script lang="ts">
    import { Label } from '@smui/list'
    import SegmentedButton, { Segment } from '@smui/segmented-button'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import Select, { Option } from '@smui/select'
    import Icon from '@smui/select/icon'
    import { onMount } from 'svelte'
    import TimePicker from '../../../components/TimePicker.svelte'
    import { sleep } from '../../../global'
    import { homescripts, timeUntilExecutionText, triggerMetaData } from '../main'
    import type { addAutomation } from '../main'
    import HmsSelector from '../../../components/Homescript/HmsSelector.svelte'

    // Static resource for displaying the segmented buttons
    const days: string[] = ['su', 'mo', 'tu', 'we', 'th', 'fr', 'sa']

    // Data which is dispatched as soon as the create button is pressed
    export let data: addAutomation = {
        name: '',
        description: '',
        hour: 0,
        minute: 0,
        days: [],
        homescriptId: '',
        enabled: true,
        trigger: 'cron',
        triggerInterval: 60,
    }

    let intervalBuffer = 1
    export let intervalUnit: 'seconds' | 'minutes' | 'hours' | 'days' = 'seconds'

    let mounted = false
    $: if (mounted && intervalBuffer) {
        switch (intervalUnit) {
            case 'seconds':
                data.triggerInterval = intervalBuffer
                break
            case 'minutes':
                data.triggerInterval = intervalBuffer * 60
                break
            case 'hours':
                data.triggerInterval = intervalBuffer * 60 * 60
                break
            case 'days':
                data.triggerInterval = intervalBuffer * 60 * 60 * 24
                break
        }
    }

    // Selected days are stored in a string[] instead of the final number[] representation
    // Is transformed into the final representation when the event is dispatched
    export let selectedDays: string[] = []

    let runsNow = false
    let timeUntilString = ''

    // Recursive function which updates the `timeUntilString` every 100ms
    // Also updates the `runsNow` boolean
    function updateTimeUntilExecutionText() {
        timeUntilString = timeUntilExecutionText(new Date(), data.hour, data.minute)
        runsNow = data.hour === new Date().getHours() && data.minute === new Date().getMinutes()
        setTimeout(updateTimeUntilExecutionText, 500)
    }

    function initInterval() {
        if (data.triggerInterval === undefined) {
            intervalBuffer = 1
        } else if (data.triggerInterval % (60 * 60 * 24) === 0) {
            intervalUnit = 'days'
            intervalBuffer = data.triggerInterval / (60 * 60 * 24)
        } else if (data.triggerInterval % (60 * 60) === 0) {
            intervalUnit = 'hours'
            intervalBuffer = data.triggerInterval / (60 * 60)
        } else if (data.triggerInterval % 60 === 0) {
            intervalUnit = 'minutes'
            intervalBuffer = data.triggerInterval / 60
        } else {
            intervalUnit = 'seconds'
            intervalBuffer = data.triggerInterval
        }
    }

    // Allows initially set days
    onMount(() => {
        selectedDays = data.days.map(d => days[d])
        initInterval()
        mounted = true
        // Start the time until updater
        updateTimeUntilExecutionText()
    })
</script>

<div class="container">
    <!-- Left -->
    <div class="left">
        <div class="trigger">
            <!-- Trigger -->
            <div class="trigger-mode">
                <span class="text-hint">Trigger</span>
                <Select withLeadingIcon={true} variant="outlined" bind:value={data.trigger}>
                    <svelte:fragment slot="leadingIcon">
                        <Icon class="material-icons">{triggerMetaData[data.trigger].icon}</Icon>
                    </svelte:fragment>
                    {#each Object.keys(triggerMetaData) as trigger}
                        <Option value={trigger}>{triggerMetaData[trigger].name}</Option>
                    {/each}
                </Select>
            </div>
        </div>

        <!-- Names and Text -->
        <div class="text">
            <span class="text-hint">Name and description of the automation</span>
            <Textfield
                bind:value={data.name}
                input$maxlength={30}
                label="Name"
                required
                style="width: 100%;"
                helperLine$style="width: 100%;"
            >
                <svelte:fragment slot="helper">
                    <CharacterCounter>0 / 30</CharacterCounter>
                </svelte:fragment>
            </Textfield>
            <Textfield
                bind:value={data.description}
                label="Description"
                style="width: 100%;"
                helperLine$style="width: 100%;"
            />
        </div>

        {#if data.trigger === 'cron'}
            <div class="trigger-cron-settings">
                <!-- Time -->
                <div class="trigger-cron-settings__time" class:disabled={data.trigger !== 'cron'}>
                    <span class="text-hint">Time when the automation runs</span>
                    <TimePicker
                        bind:hour={data.hour}
                        bind:minute={data.minute}
                        helperText={runsNow ? 'Right now' : timeUntilString}
                        invalidText={''}
                    />
                </div>
                <!-- Days -->
                <div class="trigger-cron-settings__days" class:disabled={data.trigger !== 'cron'}>
                    <span class="text-hint">Days on which the automation should run</span>
                    <SegmentedButton segments={days} let:segment bind:selected={selectedDays}>
                        <Segment
                            {segment}
                            on:click={async () => {
                                await sleep(1)
                                data.days = selectedDays.map(d => days.indexOf(d))
                                data = data
                            }}
                        >
                            <Label>{segment}</Label>
                        </Segment>
                    </SegmentedButton>
                </div>
            </div>
        {:else if data.trigger === 'interval'}
            <br />
            <span class="text-hint">Interval Settings</span>
            <div class="trigger-interval-settings">
                <div class="trigger-interval-settings__input">
                    <!-- Interval duration-->
                    <Textfield
                        bind:value={intervalBuffer}
                        label={intervalUnit}
                        invalid={intervalBuffer <= 0 || data.triggerInterval > 60 * 60 * 24 * 365}
                        type="number"
                        input$step="1"
                    />
                    <span
                        class="text-hint"
                        class:text-disabled={intervalUnit === 'seconds' || intervalBuffer <= 0}
                        style="font-size: .8rem;">= {data.triggerInterval}s</span
                    >
                </div>
                <!-- TODO: implement invalid checks in other places too-->
                <div class="trigger-interval-settings__unit">
                    <!-- Duration unit selection -->
                    <SegmentedButton
                        segments={['seconds', 'minutes', 'hours', 'days']}
                        let:segment
                        singleSelect
                        bind:selected={intervalUnit}
                    >
                        <Segment {segment}>
                            <Label>{segment}</Label>
                        </Segment>
                    </SegmentedButton>
                </div>
            </div>
        {/if}
    </div>

    <!-- Right -->
    <div class="right">
        <div class="hms">
            <span class="text-hint">The Homescript to be executed</span>
            <HmsSelector bind:selection={data.homescriptId} homescripts={$homescripts} />
        </div>
    </div>
</div>

<style lang="scss">
    @use '../../../mixins' as *;

    .container {
        display: flex;
        flex-wrap: wrap;

        @include not-widescreen {
            flex-direction: column;
        }
    }
    .left,
    .right {
        @include widescreen {
            width: 50%;
            box-sizing: border-box;
            padding: 0 1rem;
        }

        @include not-widescreen {
            width: 99%;
        }
    }

    .text {
        margin-top: 2rem;
    }

    .trigger-interval-settings {
        display: flex;
        justify-content: space-between;
        gap: 1rem;

        @include mobile {
            flex-direction: column;
            justify-content: flex-start;
        }
    }

    .trigger-cron-settings {
        display: flex;
        justify-content: space-between;
        gap: 1rem;

        @include mobile {
            flex-direction: column;
            justify-content: flex-start;
        }

        &__days {
            display: flex;
            flex-direction: column;
            gap: 0.3rem;
            margin-top: 2rem;

            &.disabled {
                user-select: none;
                pointer-events: none;
                opacity: 40%;
            }
        }

        &__time {
            margin-top: 2rem;
            display: flex;
            flex-direction: column;
            gap: 0.3rem;
            transition: 0.2s opacity;

            &.disabled {
                user-select: none;
                pointer-events: none;
                opacity: 40%;
            }
        }
    }

    .trigger {
        display: flex;
        align-items: center;
        gap: 2.5rem;
        flex-wrap: wrap;

        @include mobile {
            gap: 0;
        }
    }

    .trigger-mode {
        margin-top: 0.5rem;
    }

    .hms,
    .trigger-mode {
        display: flex;
        flex-direction: column;
        gap: 0.3rem;
    }

    .hms {
        margin: 1rem 0;

        @include widescreen {
            height: 28rem;
        }
    }

    .text {
        width: 90%;
    }
</style>
