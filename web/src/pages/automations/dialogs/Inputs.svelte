<!-- Contains the Input elements used by `AddAutomation` and `EditAutomation` -->
<script lang="ts">
    import { Label } from "@smui/list";
    import SegmentedButton, { Segment } from "@smui/segmented-button";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import { onMount } from "svelte";
    import TimePicker from "../../../components/TimePicker.svelte";
    import { sleep } from "../../../global";
    import { addAutomation, homescripts } from "../main";
    import HmsSelector from "../../../components/Homescript/HmsSelector.svelte";

    // Static resource for displaying the segmented buttons
    const days: string[] = ["su", "mo", "tu", "we", "th", "fr", "sa"];

    // Data which is dispatched as soon as the create button is pressed
    export let data: addAutomation;

    // Selected days are stored in a string[] instead of the final number[] representation
    // Is transformed into the final representation when the event is dispatched
    export let selectedDays: string[] = [];

    // Allows initially set days
    onMount(() => {
        selectedDays = data.days.map((d) => days[d]);
    });
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
                <Segment
                    {segment}
                    on:click={async () => {
                        await sleep(1);
                        data.days = selectedDays.map((d) => days.indexOf(d));
                        data = data;
                    }}
                >
                    <Label>{segment}</Label>
                </Segment>
            </SegmentedButton>
        </div>

        <div class="timing">
            <!-- Timing Mode -->
            <div class="timing-mode">
                <span class="text-hint">Timing mode</span>
                <SegmentedButton
                    segments={["normal", "sunrise", "sunset"]}
                    let:segment
                    singleSelect
                    bind:selected={data.timingMode}
                >
                    <Segment {segment}>
                        <Label>{segment}</Label>
                    </Segment>
                </SegmentedButton>
            </div>

            <!-- Time -->
            <div class="time" class:disabled={data.timingMode !== "normal"}>
                <span class="text-hint">Time when the automation runs</span>
                <TimePicker
                    bind:hour={data.hour}
                    bind:minute={data.minute}
                    helperText={"Time"}
                    invalidText={"error"}
                />
            </div>
        </div>
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
    @use "../../../mixins" as *;

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
    .days {
        margin-top: 2rem;
    }

    .time {
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

    .timing {
        display: flex;
        align-items: center;
        gap: 2.5rem;
        flex-wrap: wrap;

        @include mobile {
            gap: 0;
        }
    }

    .timing-mode {
        margin-top: 0.5rem;
    }

    .hms,
    .timing-mode,
    .days {
        display: flex;
        flex-direction: column;
        gap: 0.3rem;
    }
    .text {
        width: 90%;
    }
</style>
