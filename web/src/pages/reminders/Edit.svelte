<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import Dialog,{ Actions,Content,Header,Title } from '@smui/dialog'
    import IconButton from '@smui/icon-button'
    import Inputs from './Inputs.svelte'

    export let modify: Function

    let open = false
    let response = 'Nothing yet.'

    export let inputName = ''
    export let inputDescription = ''
    export let inputDueDate: Date
    export let selectedPriority: number

    function closeHandler(e: CustomEvent<{ action: string }>) {
        switch (e.detail.action) {
            case 'close':
                response = 'Closed without response.'
                break
            case 'reject':
                response = 'Rejected.'
                break
            case 'accept':
                response = 'Accepted.'
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
        <Inputs onSubmit={modify} submitLabel={'modify'} />
    </Content>
    <Actions>
        <Button action="modify">
            <Label>Modify</Label>
        </Button>
        <Button action="cancel" defaultAction>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>

<Button on:click={() => (open = true)}>
    <Label>Modify Reminder</Label>
</Button>

<pre class="status">Response: {response}</pre>
