<script lang="ts">
    import FormField from "@smui/form-field";
    import { Label } from "@smui/list";
    import SegmentedButton, { Segment } from "@smui/segmented-button";
    import Switch from "@smui/switch/src/Switch.svelte";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import { onMount } from "svelte";
    import TimePicker from "../../../components/TimePicker.svelte";
    import { sleep } from "../../../global";
    import type { homescriptData } from "../main";
    import HmsSelector from "./HmsSelector.svelte";

    // Data which is dispatched as soon as the create button is pressed
    export let data: homescriptData;
</script>

<div class="container">
        <!-- Names and Text -->
        <div class="text">
            <span class="text-hint">Name and description of the Homescript</span
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
        <div class="toggles">
            <div class="toggles__item">
                <Switch bind:checked={data.schedulerEnabled} />
                <span class="text-hint">
                    Automation selection {data.schedulerEnabled
                        ? "shown"
                        : "hidden"}
                </span>
            </div>
            <div class="right__toggles__item">
                <Switch bind:checked={data.quickActionsEnabled} />
                <span class="text-hint">
                    Quick actions selection {data.quickActionsEnabled
                        ? "shown"
                        : "hidden"}
                </span>
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

    .text {
        width: 100%;
    }
</style>
