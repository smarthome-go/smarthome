<script lang="ts">
    import Button from '@smui/button'
    import SegmentedButton,{ Label,Segment } from '@smui/segmented-button'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import HelperText from '@smui/textfield/helper-text'
    import DatePicker from '../../components/DatePicker.svelte'

    const priorities = ['Low', 'Normal', 'Medium', 'High', 'Urgent'] // Priorities for translating the current choice to a number

    let datePicker: DatePicker // Date picker component
    const defaultDate = new Date() // Used to check if the date is `dirty`

    /** Bindable data variables */
    export let inputName = ''
    export let inputDescription = ''
    export let inputDueDate = defaultDate
    export let selectedPriority = 'Normal'

    /** Customization*/
    export let showButtons = true // Used in the modification popup, buttons not needed there

    /** Dirty detection and cleaning */
    let dirty = false
    $: dirty = nameDirty || descriptionDirty || dueDateDirty || priorityDirty

    let nameDirty = false
    let descriptionDirty = false
    let dueDateDirty = false
    let priorityDirty = false

    $: dueDateDirty = inputDueDate != defaultDate // Default used as comparison instead of `new Date`
    $: priorityDirty = selectedPriority != 'Normal' // Normal is the default priority

    export function clear() {
        inputName = ''
        inputDescription = ''
        selectedPriority = 'Normal'
        inputDueDate = defaultDate
        datePicker.clear()

        nameDirty = false
        descriptionDirty = false
    }

    /** Date picker validation: if the due date is more than 30 days in the past, it is invalid */
    const now = new Date()
    const thirtyDaysInMs = 30 * 24 * 60 * 60 * 1000
    let datePickerInvalid = false
    $: {
        if (inputDueDate !== undefined && inputDueDate !== null)
            datePickerInvalid =
                now.getTime() - inputDueDate.getTime() + 86400000 >=
                thirtyDaysInMs // The `8.64e+7` is for adding one extra day to the currently selected date
    }

    export let onSubmit: Function // Callback to be executed if the create / submit button is used
</script>

<!-- Name -->
<div id="name">
    <Textfield
        style="width: 100%;"
        helperLine$style="width: 100%;"
        bind:value={inputName}
        bind:dirty={nameDirty}
        label="Name"
        input$maxlength={100}
    >
        <CharacterCounter slot="helper">0 / 100</CharacterCounter>
    </Textfield>
</div>

<!-- Description -->
<div id="description">
    <Textfield
        style="width: 100%;"
        helperLine$style="width: 100%;"
        textarea
        bind:value={inputDescription}
        bind:dirty={descriptionDirty}
        label="Description"
        input$rows={5}
    >
        <HelperText slot="helper"
            >Describe which task you want to accomplish</HelperText
        >
    </Textfield>
</div>

<div id="priority-duedate">
    <!-- Priority -->
    <div id="priority">
        <p class="text-hint">Priority</p>
        <SegmentedButton
            segments={priorities}
            let:segment
            singleSelect
            bind:selected={selectedPriority}
        >
            <Segment {segment}>
                <Label>{segment}</Label>
            </Segment>
        </SegmentedButton>
    </div>

    <!-- Due Date -->
    <div id="duedate">
        <p class="text-hint">Due Date</p>
        <DatePicker
            invalidText={'Due Date is more than a month in the past'}
            bind:this={datePicker}
            helperText={'The date on which the task should be completed'}
            bind:value={inputDueDate}
            bind:invalid={datePickerInvalid}
        />
    </div>
</div>

<!-- Submit / Cancel button -->
{#if showButtons}
    <div id="buttons" class="align">
        <Button
            on:click={async () => {
                await onSubmit(
                    inputName,
                    inputDescription,
                    priorities.indexOf(selectedPriority),
                    inputDueDate
                )
                clear()
            }}
            disabled={inputName.length === 0 ||
                datePickerInvalid}
            touch
            variant="raised"
        >
            <Label>Create</Label>
        </Button>
        <Button disabled={!dirty} on:click={clear} touch>
            <Label>Cancel</Label>
        </Button>
    </div>
{/if}

<style lang="scss">
    @use '../../mixins' as *;

    #description {
        margin-top: 1rem;
        :global(.mdc-text-field__resizer) {
            resize: none;
        }
    }

    #duedate {
        margin-top: .3rem;
    }

    #priority-duedate {
        margin-top: .5rem;
        display: flex;
        gap: 1rem;
        flex-wrap: wrap;

        p {
            font-size: .7rem;
            margin: .2rem 0;
        }
    }

    .align {
        display: flex;
        align-items: center;
        gap: 1rem;

        @include mobile {
            gap: 0.7rem;
        }
    }
</style>
