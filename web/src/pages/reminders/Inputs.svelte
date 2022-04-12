<script lang="ts">
    import Button from '@smui/button'
    import SegmentedButton,{ Label,Segment } from '@smui/segmented-button'
    import Textfield from '@smui/textfield'
    import CharacterCounter from '@smui/textfield/character-counter'
    import HelperText from '@smui/textfield/helper-text'
    import DatePicker from '../../components/DatePicker.svelte'

    // Date picker
    let datePicker: DatePicker
    const defaultDate = new Date()

    // Priorities
    const priorities = ['Low', 'Normal', 'Medium', 'High', 'Urgent']

    // Bindable data variables
    export let inputName = ''
    export let inputDescription = ''
    export let inputDueDate = defaultDate
    export let selectedPriority = 'Normal'

    // Customization and modes
    export let submitLabel = 'submit'

    // Dirty-variables and clearing
    let dirty = false
    $: dirty = nameDirty || descriptionDirty || dueDateDirty || priorityDirty

    let nameDirty = false
    let descriptionDirty = false
    let dueDateDirty = false
    let priorityDirty = false

    $: dueDateDirty = inputDueDate != defaultDate
    $: priorityDirty = selectedPriority != 'Normal'

    export function clear() {
        inputName = ''
        inputDescription = ''
        selectedPriority = 'Normal'
        inputDueDate = defaultDate
        datePicker.clear()

        nameDirty = false
        descriptionDirty = false
    }

    export let onSubmit: Function
</script>

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
<br />
<br />
<DatePicker
    bind:this={datePicker}
    label={'Due Date'}
    bind:value={inputDueDate}
/>
<br />
<!-- Create and cancel buttons -->
<div class="align">
    <Button
        on:click={async() => {
            await onSubmit(inputName, inputDescription, priorities.indexOf(selectedPriority), inputDueDate)
            clear()
        }}
        disabled={inputName.length === 0 || !dueDateDirty}
        touch
        variant="raised"
    >
        <Label>{submitLabel}</Label>
    </Button>
    <Button disabled={!dirty} on:click={clear} touch>
        <Label>Cancel</Label>
    </Button>
</div>

<style lang="scss">
    @use '../../mixins' as *;

    #description {
        margin-top: 1rem;
        :global(.mdc-text-field__resizer) {
            resize: none;
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
