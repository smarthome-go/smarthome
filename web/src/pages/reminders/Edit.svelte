<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Header,InitialFocus,Title } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import Inputs from './Inputs.svelte'

    export let modify: (_name: string, _description: string, _priority: number, _dueDate: Date) => Promise<void>

    let open = false

    const priorities = ['Low', 'Normal', 'Medium', 'High', 'Urgent']

    export let inputName = ''
    export let inputDescription = ''
    export let inputDueDate: Date
    export let selectedPriority = 0
    let priority = priorities[selectedPriority]

    $: selectedPriority = priorities.indexOf(priority)

    // Values when the dialog is opened, used for reverting the changes on cancel
    let nameBefore: string
    let descriptionBefore: string
    let dueDateBefore: Date
    let priorityBefore: number

    // Date Picker validation
    const now = new Date()
    const thirtyDaysInMs = 30 * 24 * 60 * 60 * 1000
    let datePickerInvalid = false
    $: {
        if (inputDueDate !== undefined && inputDueDate !== null)
            datePickerInvalid =
                now.getTime() - inputDueDate.getTime() + 86400000 >=
                thirtyDaysInMs // The `8.64e+7` is for adding one extra day
    }

    function closeHandler(e: CustomEvent<{ action: string }>) {
        switch (e.detail.action) {
            case 'modify':
                modify(
                    inputName,
                    inputDescription,
                    selectedPriority,
                    inputDueDate
                )
                nameBefore = inputName
                descriptionBefore = inputDescription
                priorityBefore = selectedPriority
                dueDateBefore = inputDueDate
                break
            default:
                // Reset all values to their original state
                inputName = nameBefore
                inputDescription = descriptionBefore
                selectedPriority = priorityBefore
                inputDueDate = dueDateBefore
                break
        }
    }
</script>

<Dialog
    bind:open
    fullscreen
    aria-labelledby="title"
    aria-describedby="content"
    on:SMUIDialog:closed={closeHandler}
>
    <Header>
        <Title id="title">Modify Reminder</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="content">
        <Inputs
            bind:inputName
            bind:inputDescription
            bind:inputDueDate
            bind:selectedPriority={priority}
            onSubmit={modify}
            showButtons={false}
        />
    </Content>
    <Actions>
        <Button defaultAction action="cancel">
            <Label>Cancel</Label>
        </Button>
        <Button use={[InitialFocus]} disabled={datePickerInvalid} action="modify">
            <Label>Modify</Label>
        </Button>
    </Actions>
</Dialog>

<IconButton
    class="material-icons"
    on:click={async () => {
        open = true
        nameBefore = inputName
        descriptionBefore = inputDescription
        priorityBefore = selectedPriority
        dueDateBefore = inputDueDate
    }}
    title="Edit Reminder">edit</IconButton
>
