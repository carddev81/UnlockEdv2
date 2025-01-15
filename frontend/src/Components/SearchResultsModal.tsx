import { KiwixChannel, PaginationMeta } from '@/common';
import { CloseX } from './inputs/CloseX';
import { forwardRef, ReactNode } from 'react';
import Pagination from './Pagination';
//import Pagination from './Pagination';

interface SearchResultsModalProps {
    meta: PaginationMeta;
    channel: KiwixChannel; // The entire channel data with items and metadata
    children?: ReactNode; // for loading..???
    onItemClick: (link: string) => void; 
    onPageChange: (page: number, perPage: number) => void; //added this for pagination
}

const SearchResultsModal = forwardRef<
    HTMLDialogElement,
    SearchResultsModalProps
>(function SearchResultsModal({ meta, channel, children, onItemClick, onPageChange }, ref) {
    // const [perPage, setPerPage] = useState(10);
    // const [pageQuery, setPageQuery] = useState(1);

    const parseBoldTags = (text: string) => {
        const regex = /<b>(.*?)<\/b>/g; 
        const parts = text.split(regex); 
        const elements = parts.map((part, index) =>
            index % 2 === 1 ? <strong key={index}>{part}</strong> : part
        );
        return <>{elements}</>;
    }
    const parseNumber = (num: string) => {
            return Number(num.replace(/,/g, ""));
    }
    console.log(channel.start_index);
    console.log(channel.total_results);
    console.log(channel.items_per_page);
    const currentEndResultsPage = Math.min(Number(parseNumber(channel.start_index)) + Number(channel.items_per_page) - 1, parseNumber(channel.total_results))
    const currentPageDisplay = `Results ${channel.start_index}-${currentEndResultsPage} of ${channel.total_results}`
    return (
        // inner-background
        <dialog
            ref={ref}
            className="modal fixed inset-0 w-full h-full bg-black/50 flex justify-center items-center overflow-visible"
        >
            <div
                className="modal-box max-w-6xl w-full max-h-[700px] overflow-hidden rounded-lg shadow-lg relative"
            >
                <div className="sticky top-0 bg-white z-50 pb-4 border-b border-gray-300">
                        {/* Title and Description */}
                    <div>
                    <div className="flex items-center gap-4 px-4">
                        <div className="flex items-center gap-2 border-b-2">
                            <figure className="w-[48px] h-[48px] bg-cover">
                                <img
                                    src={channel.thumbnail_url ?? ''}
                                    alt="placeholder"
                                    className="w-full h-full object-cover"
                                />
                            </figure>
                            <h3 className="w-auto body my-auto">Add title here!!!</h3>
                        </div>
                            <div className="flex flex-col">
                                <h2 className="text-2xl font-bold">{channel.description}</h2>
                                <p className="text-sm text-gray-600">{currentPageDisplay}</p>
                            </div>
                        </div>
                        {/* added close button to search modal; also TODO custom cards will be here!!!! */}
                        <CloseX
                            close={() =>
                                (
                                    ref as React.RefObject<HTMLDialogElement>
                                ).current?.close()
                            }
                        />
                    </div>
                </div>
                <div className="overflow-y-auto max-h-[450px] px-4 pb-4">
                    <div className="flex flex-col gap-4"> 
                        {channel.items?.map((item, index) => (
                            <div
                                key={index}
                                className="card bg-grey-1 p-4 rounded-lg shadow-md cursor-pointer"
                                onClick={() => onItemClick(item.link)}
                            >
                                <h3 className="font-bold text-lg mb-2">{item.title}</h3>
                                <p className="text-sm text-gray-700 mb-2">{parseBoldTags(item.description)}</p>
                                <p className="text-sm text-gray-600">Book: {item.book}</p>
                                <p className="text-sm text-gray-600">Word Count: {item.word_count}</p>
                            </div>
                        ))}
                    </div>
                    <div className="flex justify-center mt-4">
                    <Pagination
                        meta={meta}
                        setPage={(page) =>
                            onPageChange(page, Number(channel.items_per_page))
                        }
                        setPerPage={(perPage) =>
                            onPageChange(
                                1, // Reset to the first page when perPage changes
                                perPage
                            )
                        }
                    />
                     </div>
                </div>
                {children}
            </div>
        </dialog>
    );
});

export default SearchResultsModal;