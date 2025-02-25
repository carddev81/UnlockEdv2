import useSWR from 'swr';
import { AxiosError } from 'axios';
import {
    OpenContentResponse,
    ResidentEngagementProfile,
    ServerResponseOne
} from '@/common';
import EngagementRateGraph from '@/Components/EngagementRateGraph';
import { ResponsiveContainer } from 'recharts';
import StatsCard from '@/Components/StatsCard';
import { UserCircleIcon } from '@heroicons/react/24/outline';
import { useParams } from 'react-router-dom';
import ClampedText from '@/Components/ClampedText';

const StudentProfile = () => {
    const { user_id } = useParams<{ user_id: string }>();
    const { data, error, isLoading } = useSWR<
        ServerResponseOne<ResidentEngagementProfile>,
        AxiosError
    >(`/api/users/${user_id}/profile`);
    const metrics = data?.data;

    return (
        <div className="overflow-x-hidden px-5 pb-4">
            {error && <div>Error loading data</div>}
            {!data || (isLoading && <div>Loading...</div>)}
            {data && metrics && (
                <>
                    <div className="flex flex-row gap-6 items-stretch">
                        <div className="w-[270px] h-[240px] flex flex-col gap-4">
                            <div className="card card-row-padding overflow-hidden text-med flex-1 h-full">
                                <div className="justify-items-center">
                                    <UserCircleIcon className="w-[64px] h-[64px]" />
                                </div>
                                <div className="">
                                    {(() => {
                                        const { name_first, name_last } =
                                            metrics.login_engagement
                                                .peak_login_times[0];
                                        return (
                                            <div className="text-sm mt-2">
                                                {name_first} {name_last}{' '}
                                            </div>
                                        );
                                    })()}
                                    <div className="text-sm mt-2">
                                        <span className="font-semibold  justify-self-start">
                                            Username
                                        </span>
                                        {' : '}
                                        {
                                            metrics.login_engagement
                                                .peak_login_times[0].username
                                        }
                                    </div>
                                    <div className="text-sm mt-2">
                                        <span className="font-semibold">
                                            Joined :
                                        </span>{' '}
                                        {new Date(
                                            metrics.activity_engagement.first_active_date
                                        ).toLocaleDateString('en-US')}
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div className="flex-1 h-[240px] flex flex-col gap-4">
                            <div className="card card-row-padding overflow-hidden">
                                <h1 className="">
                                    {' '}
                                    {metrics.login_engagement
                                        .peak_login_times[0].name_first +
                                        " 's recent Activity"}
                                </h1>
                                <div className=" items-stretch">
                                    <div className="h-[240px] overflow-visible">
                                        <ResponsiveContainer
                                            className="h-full p-7"
                                            width="100%"
                                            height="100%"
                                            debounce={500}
                                        >
                                            <EngagementRateGraph
                                                peak_login_times={
                                                    metrics?.login_engagement
                                                        .peak_login_times ?? []
                                                }
                                                viewType={'daily'}
                                            />
                                        </ResponsiveContainer>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div className="grid grid-cols-3 gap-4 mb-6 mt-6">
                        <StatsCard
                            title="Days Active"
                            number={metrics.activity_engagement.total_hours_active_monthly.toFixed(
                                2
                            )}
                            label="Days Active This Month"
                            tooltip="Total number of days resident has been active in UnlockedEd"
                        />
                        <StatsCard
                            title="Average Hours"
                            number={
                                metrics.activity_engagement
                                    .total_hours_active_weekly
                            }
                            label={'AVG Hours PER Week'}
                            tooltip={
                                'Average number of hours resident is logged in to UnlockedEd'
                            }
                        />
                        <StatsCard
                            title="Total Hours"
                            number={metrics.activity_engagement.total_hours_engaged.toFixed(
                                2
                            )}
                            label={'Total Hours This Week'}
                            tooltip={
                                'Total number of hours resident was logged in to UnlockedEd this week'
                            }
                        />
                    </div>
                    <div className="grid grid-cols-3 gap-3 mb-6 mt-6">
                        <div>
                            <h2 className="text-center">
                                Top 5 Most Viewed Libraries
                            </h2>
                            <table className="table-2 mb-4">
                                <thead>
                                    <tr className="grid-col-3">
                                        <th className="justify-self-start">
                                            Library Name
                                        </th>
                                        <th># Hours Watching</th>
                                        <th className="justify-self-end">
                                            Is Featured
                                        </th>
                                    </tr>
                                </thead>
                                <tbody className="flex flex-col gap-4 mt-4">
                                    {metrics.top_libraries.map(
                                        (items: OpenContentResponse) => {
                                            return (
                                                <tr>
                                                    <td className="justify-self-start">
                                                        <img
                                                            className="h-8 mx-auto object-contain"
                                                            src={
                                                                items.thumbnail_url ??
                                                                ''
                                                            }
                                                        />
                                                        <ClampedText
                                                            as="h3"
                                                            lines={1}
                                                            className="my-auto w-full body font-normal text-left"
                                                        >
                                                            {items.title ??
                                                                'Untitled'}
                                                        </ClampedText>
                                                    </td>
                                                    <td>{items.total_hours}</td>
                                                    <td className="justify-self-end">
                                                        <input
                                                            name={'is_featured'}
                                                            type="checkbox"
                                                            className="checkbox"
                                                            checked={
                                                                items.is_featured
                                                            }
                                                        />
                                                    </td>
                                                </tr>
                                            );
                                        }
                                    )}
                                </tbody>
                            </table>
                        </div>
                        <div></div>
                        <div>
                            <h2 className="text-center">
                                Top 5 Recently Watched Videos
                            </h2>
                            <table className="table-2 mb-4">
                                <thead>
                                    <tr className="grid-col-2">
                                        <th>Rank</th>
                                        <th>Title</th>
                                    </tr>
                                </thead>
                                <tbody className="flex flex-col gap-4 mt-4">
                                    {metrics.recent_videos.length > 0 ? (
                                        metrics.recent_videos.map(
                                            (
                                                items: OpenContentResponse,
                                                index: number
                                            ) => {
                                                return (
                                                    <tr>
                                                        <td>{index + 1}</td>
                                                        <td>
                                                            <img
                                                                className="h-8 mx-auto object-contain"
                                                                src={
                                                                    '/src/assets/react.svg'
                                                                }
                                                            />
                                                            <ClampedText
                                                                as="h3"
                                                                lines={1}
                                                                className="my-auto w-full body font-normal text-left"
                                                            >
                                                                {items.title ??
                                                                    'Untitled'}
                                                            </ClampedText>
                                                        </td>
                                                    </tr>
                                                );
                                            }
                                        )
                                    ) : (
                                        <tr>
                                            <td>{1}</td>
                                            <td>
                                                <img
                                                    className="h-8 mx-auto object-contain"
                                                    src={
                                                        '/src/assets/react.svg'
                                                    }
                                                />
                                                <ClampedText
                                                    as="h3"
                                                    lines={1}
                                                    className="my-auto w-full body font-normal text-left"
                                                >
                                                    {'Untitled'}
                                                </ClampedText>
                                            </td>
                                        </tr>
                                    )}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </>
            )}
        </div>
    );
};

export default StudentProfile;
