<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Header,Title } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import Inputs from './Inputs.svelte'

    export let modify: Function

    let open = false

    const priorities = ['Low', 'Normal', 'Medium', 'High', 'Urgent']

    export let inputName = ''
    export let inputDescription = ''
    export let inputDueDate: Date
    export let selectedPriority: number
    let priority = priorities[selectedPriority]

    $: selectedPriority = priorities.indexOf(priority)

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
            case 'close' || 'cancel':
                break
            case 'modify':
                modify(
                    inputName,
                    inputDescription,
                    selectedPriority,
                    inputDueDate
                )
                break
        }
    }
</script>

<Dialog
    bind:open
    fullscreen
    aria-labelledby="fullscreen-title"
    aria-describedby="fullscreen-content"
    on:SMUIDialog:closed={closeHandler}
>
    <Header>
        <Title id="fullscreen-title">Modify Reminder</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="fullscreen-content">
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
        <Button disabled={datePickerInvalid} action="modify">
            <Label>Modify</Label>
        </Button>
        <Button action="cancel" defaultAction>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>

<IconButton
    class="material-icons"
    on:click={() => (open = true)}
    title="Edit Reminder">edit</IconButton
>