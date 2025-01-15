import { useEffect, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import Error from '@/Pages/Error';
import API from '@/api/api';
import { KiwixChannel, Library, PaginationMeta, ServerResponseMany, ServerResponseOne } from '@/common';
import { usePathValue } from '@/Context/PathValueCtx';
import LibrarySearchBar from '@/Components/inputs/LibrarySearchBar';
import SearchResultsModal from '@/Components/SearchResultsModal';

export default function LibraryViewer() {
    const { id: libraryId } = useParams();
    const [src, setSrc] = useState<string>('');
    const [error, setError] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const { setPathVal } = usePathValue();

    //search stuff
    const [searchTerm, setSearchTerm] = useState('');
    //const searchQuery = useDebounceValue(searchTerm, 300);
    const [isModalLoading, setIsModalLoading] = useState(false); // Loading state for modal
    const [searchResults, setSearchResults] = useState<KiwixChannel | null>(
        null
    ); // Results state

        const [meta, setMeta] = useState<PaginationMeta>({
            current_page: 1,
            per_page: 10,
            total: 0,
            last_page: 0
        });
    const modalRef = useRef<HTMLDialogElement>(null);
    const openModal = () => modalRef.current?.showModal();
    const closeModal = () => modalRef.current?.close();
    const handleItemClick = (link: string) => {
        setSrc(link); // Update iframe source
        closeModal(); // Close the modal
    };
    useEffect(() => {
        const fetchLibraryData = async () => {
            setIsLoading(true);
            try {
                const resp = (await API.get(
                    `libraries/${libraryId}`
                )) as ServerResponseOne<Library>;
                if (resp.success) {
                    setPathVal([
                        { path_id: ':library_name', value: resp.data.title }
                    ]);
                }
                const response = await fetch(
                    `/api/proxy/libraries/${libraryId}/`
                );
                if (response.ok) {
                    setSrc(response.url);
                } else if (response.status === 404) {
                    setError('Library not found');
                } else {
                    setError('Error loading library');
                }
            } catch {
                setError('Error loading library');
            } finally {
                setIsLoading(false);
            }
        };
        void fetchLibraryData();
        return () => {
            sessionStorage.removeItem('tag');
        };
    }, [libraryId]);

    // Handle Search Logic
    const handleSearch: () => void = () => {
        void (async () => {
            openModal();
            setIsModalLoading(true);
            try {
                const response = (await API.get(
                    `libraries/${libraryId}/search?pattern=${searchTerm}`
                )) as ServerResponseMany<KiwixChannel>;
                setMeta(response.meta)
                setSearchResults(response.data[0]);
            } catch {
                setSearchResults(null);
            } finally {
                setIsModalLoading(false);
            }
        })();
    };
    const handlePageChange = (page: number, perPage: number): void => {
        void (async () => {
            setIsModalLoading(true);
            try {
                const response = (await API.get(
                    `libraries/${libraryId}/search?pattern=${searchTerm}&page=${page}&per_page=${perPage}`
                )) as ServerResponseMany<KiwixChannel>;
                setMeta(response.meta)
                setSearchResults(response.data[0]);
            } catch {
                setSearchResults(null); 
            } finally {
                setIsModalLoading(false);
            }
        })();
    };    // const mockChannel: KiwixChannel = {
    //     title: 'Search Results',
    //     link: '/search?query=test',
    //     description: 'Search results for your query.',
    //     total_results: '500',
    //     start_index: '1',
    //     items_per_page: '25',
    //     items: [
    //         {
    //             title: 'Example Title 1',
    //             link: 'https://example.com/1',
    //             description: 'Description for result 1.',
    //             book: 'Book 1',
    //             word_count: '1200'
    //         },
    //         {
    //             title: 'Example Title 2',
    //             link: 'https://example.com/2',
    //             description: 'Description for result 2.',
    //             book: 'Book 2',
    //             word_count: '850'
    //         }
    //         // Add more mock items as needed
    //     ]
    // };

    return (
        <div>
            <div className="px-8 pb-4">
                <div className="flex items-center gap-4 mb-4">
                    <h1 className="text-2xl font-bold">Library Viewer</h1>
                    <LibrarySearchBar
                        searchTerm={searchTerm}
                        onSearchClick={handleSearch}
                        changeCallback={setSearchTerm}
                    />
                    <SearchResultsModal
                        meta={meta}
                        onItemClick={handleItemClick}
                        onPageChange={handlePageChange}
                        channel={
                            searchResults ?? {
                                title: 'Loading...',
                                link: '',
                                thumbnail_url: '',
                                description: '',
                                total_results: '',
                                start_index: '',
                                items_per_page: '',
                                items: []
                            }
                        }
                        ref={modalRef}
                    >
                        {isModalLoading && (
                            <div className="flex justify-center items-center h-40">
                                <span className="loading loading-spinner loading-lg"></span>
                                <p className="ml-4 text-lg">
                                    Loading Results...
                                </p>
                            </div>
                        )}
                    </SearchResultsModal>
                </div>
                <div className="w-full pt-4 justify-center">
                    {isLoading ? (
                        <div className="flex h-screen gap-4 justify-center content-center">
                            <span className="my-auto loading loading-spinner loading-lg"></span>
                            <p className="my-auto text-lg">Loading...</p>
                        </div>
                    ) : src != '' ? (
                        <iframe
                            sandbox="allow-scripts allow-same-origin allow-modals allow-popups"
                            className="w-full h-screen pt-4"
                            id="library-viewer-iframe"
                            src={src}
                        />
                    ) : (
                        error && <Error />
                    )}
                </div>
            </div>
        </div>
    );
}
