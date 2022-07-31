<script lang="ts">
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import { onMount } from "svelte";
    import TimePicker from "../../../components/TimePicker.svelte";
    import { ScheduleData, timeUntilExecutionText } from "../main";
    import HmsInputs from "./HMSInputs.svelte";

    // Data which is dispatched as soon as the create button is pressed
    export let data: ScheduleData;

    export let timeInvalid: boolean = false;
    let timeUntilString = "";

    // Recursive function which updates the `timeUntilString` every 100ms
    // Also updates the `timeInvalid` boolean
    function updateTimeUntilExecutionText() {
        timeUntilString = timeUntilExecutionText(
            new Date(),
            data.hour,
            data.minute
        );
        timeInvalid =
            data.hour === new Date().getHours() &&
            data.minute === new Date().getMinutes();
        setTimeout(updateTimeUntilExecutionText, 100);
    }

    // Start the time updater
    onMount(updateTimeUntilExecutionText);
</script>

<div class="container">
    <!-- Left -->
    <div class="left">
        <!-- Names and Text -->
        <div class="text">
            <span class="text-hint">Name of the schedule</span>
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
        </div>
        <div class="timing">
            <!-- Time -->
            <div class="time">
                <span class="text-hint">Time when the schedule runs</span>
                <TimePicker
                    bind:hour={data.hour}
                    bind:minute={data.minute}
                    helperText={timeUntilString}
                    invalid={timeInvalid}
                    invalidText={"The schedule can't run now"}
                />
            </div>
        </div>
    </div>
    <!-- Right -->
    <div class="right">
        <HmsInputs bind:data />
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
            padding-top: 1rem;
        }

        @include not-widescreen {
            width: 99%;
        }
    }
    .time {
        margin-top: 2rem;
        display: flex;
        flex-direction: column;
        gap: 0.3rem;
        transition: 0.2s opacity;
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

    .text {
        width: 90%;
    }
</style>
