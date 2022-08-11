<script lang="ts">
    import Button, { Icon, Label } from "@smui/button";
    import Dialog, { Actions, Content, Title } from "@smui/dialog";

    export let open = false;
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Geolocation</Title>
    <Content id="content">
        Geolocation is not required in order to use Smarthome.
        <br />
        However, specifying a geolocation will allow following services to work correctly.
        <ul>
            <li>
                Automation using your local <span
                    style="color: var(--clr-primary);">sunrise</span
                >
                and <span style="color: var(--clr-primary);">sunset</span> time
            </li>
            <li>
                Weather data via the <a
                    href="https://openweathermap.org"
                    rel="noopener noreferrer nofollow"
                    target="_blank"
                    style="color: var(--clr-primary);">OpenWeatherMap</a
                > API
            </li>
        </ul>
        <span class="text-disabled"
            >Smarthome will never collect or send any data.</span
        >
    </Content>
    <Actions>
        <Button>
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>
