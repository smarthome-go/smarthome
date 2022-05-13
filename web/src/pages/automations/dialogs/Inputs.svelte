<!-- Contains the Input elements used by `AddAutomation` and `EditAutomation` -->
<script lang="ts">
    import { Label } from '@smui/list'
    import SegmentedButton,{ Segment } from '@smui/segmented-button'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import TimePicker from '../../../components/TimePicker.svelte'
    import type { addAutomation } from '../main'
    import HmsSelector from './HmsSelector.svelte'

    // Static ressource for displaying the segmented buttons
    const days: string[] = ['su', 'mo', 'tu', 'we', 'th', 'fr', 'sa']

    // Data which is dispatched as soon as the create button is pressed
    export let data: addAutomation = {
        days: [],
        description: '',
        enabled: true,
        homescriptId: '',
        hour: 0,
        minute: 0,
        name: '',
        timingMode: 'normal',
    }

    // Selected days are stored in a string[] instead of the final number[] representation
    // Is transformed into the final representation when the event is dispatched
    let selectedDays: string[] = []

    // Transform the selected days into data that the server understands
    $: data.days = selectedDays.map((d) => days.indexOf(d))

</script>

<div class="container">
    <!-- Left -->
    <div class="left">
        <!-- Names and Text -->
        <div class="text">
            <span class="text-hint">Name and description of the automation</span
            >
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

        <!-- Days -->
        <div class="days">
            <span class="text-hint"
                >Days on which the automation should run</span
            >
            <SegmentedButton
                segments={days}
                let:segment
                bind:selected={selectedDays}
            >
                <Segment {segment}>
                    <Label>{segment}</Label>
                </Segment>
            </SegmentedButton>
        </div>

        <!-- Time -->
        <div class="time">
            <span class="text-hint">Time when the automation runs</span>
            <TimePicker
                bind:hour={data.hour}
                bind:minute={data.minute}
                helperText={'Time'}
                invalidText={'error'}
            />
        </div>
    </div>

    <!-- Right -->
    <div class="right">
        <div class="hms">
            <span class="text-hint">The Homescript to be executed</span>
            <HmsSelector bind:selection={data.homescriptId} />
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
    }
    .days,
    .time {
        margin-top: 2rem;
    }
    .hms,
    .time,
    .days {
        display: flex;
        flex-direction: column;
        gap: 0.3rem;
    }
    .text {
        width: 90%;
    }
</style>
