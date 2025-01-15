import { MagnifyingGlassIcon } from '@heroicons/react/24/solid';

export default function LibrarySearchBar({
    searchTerm,
    changeCallback,
    onSearchClick
}: {
    searchTerm: string;
    changeCallback: (arg: string) => void;
    onSearchClick: () => void;
}) {

//no error handling yet or focus set to scroll up to top of library results
    return (
        <label className="form-control">
            <div className="relative">
                <button
                    type="button"
                    onClick={onSearchClick}
                    className="absolute top-1/2 right-2 transform -translate-y-1/2 focus:outline-none hover:text-primary"
                    aria-label="Search"
                >
                    <MagnifyingGlassIcon className="w-5 h-5 text-black" />
                </button>
                <input
                    type="text"
                    placeholder="Search Libraries..."
                    className="input input-bordered w-full max-w-xs"
                    value={searchTerm}
                    onChange={(e) => changeCallback(e.target.value)}
                    autoFocus
                />
            </div>
        </label>
    );
}
