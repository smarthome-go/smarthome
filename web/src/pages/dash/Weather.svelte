<script lang="ts">
    import Box from "./Box.svelte";
    import { createSnackbar } from "../../global";
    import { onMount } from "svelte";

    const mainIcons = {
        Clear: ["cleary_day", "clear_night"],
        Rain: ["rain", "rain"],
        Clouds: ["cloud", "cloud"],
        Snow: ["ac_unit", "ac_unit"],
        Thunderstorm: ["thunderstorm", "thunderstorm"],
        Drizzle: ["cloudy_snowing", "cloudy_snowing"],
    };

    const descriptionIcons = {
        "few clouds": ["partly_cloudy_day", "partly_cloudy_night"],
        "scattered clouds": ["cloud", "cloud"],
    };

    let loading = false;
    let loaded = false;

    let data: weatherData = {
        id: 0,
        time: 0,
        weatherTitle: "",
        weatherDescription: "unknown",
        temperature: 0,
        feelsLike: 0,
        humidity: 0,
    };

    interface weatherData {
        id: number;
        time: number;
        weatherTitle: string;
        weatherDescription: string;
        temperature: number;
        feelsLike: number;
        humidity: number;
    }

    // Fetches a current (max. 5 minutes old) weather from the server
    async function loadWeatherData() {
        loading = true;
        try {
            let res = await (await fetch("/api/weather")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            // Signal that the weather has been loaded successfully
            data = res;
            loaded = true;
        } catch (err) {
            $createSnackbar(`Could not load weather: ${err}`);
        }
        loading = false;
    }

    onMount(loadWeatherData);
</script>

<Box bind:loading>
    <span slot="header">Weather</span>
    <div class="weather" slot="content">
        <div class="weather__top">
            {#if loaded}
                <i class="material-icons weather__top__icon">
                    {descriptionIcons[data.weatherDescription] !== undefined
                        ? descriptionIcons[data.weatherDescription][0]
                        : mainIcons[data.weatherTitle][0]}
                </i>
            {/if}
            <div class="weather__top__labels">
                <div class="weather__top__labels__title">
                    {data.weatherTitle}
                </div>
                <div class="weather__top__labels__description">
                    {data.weatherDescription}
                </div>
            </div>
        </div>
        <div class="weather__measurements">
            <div class="weather__measurements__measurement">
                <div class="weather__measurements__temperature">
                    {Math.round(data.temperature)}°C
                </div>
                <span class="text-hint">Temperature</span>
            </div>
            <div class="weather__measurements__measurement">
                <div class="weather__measurements__feels-like">
                    {Math.round(data.feelsLike)}°C
                </div>
                <span class="text-hint">Feels Like</span>
            </div>
            <div class="weather__measurements__measurement">
                <div class="weather__measurements__humidity">
                    {data.humidity}%
                </div>
                <span class="text-hint">Humidity</span>
            </div>
        </div>
    </div>
</Box>

<style lang="scss">
    .weather {
        &__top {
            display: flex;
            align-items: center;
            gap: 1.5rem;

            &__icon {
                font-size: 4rem;
            }

            &__labels {
                &__title {
                    font-size: 1.125rem;
                    font-weight: bold;
                }
                &__description {
                    font-size: 0.9rem;
                    color: var(--clr-text-hint);
                }
            }
        }
        &__measurements {
            margin-top: 1rem;
            display: flex;
            gap: 2rem;

            &__measurement {
                font-size: 1.1rem;

                .text-hint {
                    font-size: 0.8rem;
                }
            }
        }
    }
</style>
