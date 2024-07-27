import React from "react";
import "./PageButtons.css";

const PageButtons = ({ currentPage, totalPages, onPageChange }) => {
  const pageNumbers = [];

  const calculatePageRange = () => {
    if (totalPages <= 5) {
      return Array.from({ length: totalPages }, (_, i) => i + 1);
    } else if (currentPage <= 3) {
      return [1, 2, 3, "ellipsis", totalPages];
    } else if (currentPage >= totalPages - 2) {
      return [1, "ellipsis", totalPages - 2, totalPages - 1, totalPages];
    } else {
      return [
        1,
        "ellipsis",
        currentPage - 1,
        currentPage,
        currentPage + 1,
        "ellipsis",
        totalPages,
      ];
    }
  };

  const displayPages = calculatePageRange();

  return (
    <div className="pagination">
      <button
        className="page-button"
        onClick={() => onPageChange(currentPage - 1)}
        disabled={currentPage === 1}
      >
        &laquo;
      </button>
      {displayPages.map((pageNumber, index) => (
        <button
          key={index}
          className={`page-button ${
            currentPage === pageNumber ? "active" : ""
          } ${pageNumber === "ellipsis" ? "ellipsis" : ""}`}
          onClick={() =>
            onPageChange(
              typeof pageNumber === "number" ? pageNumber : currentPage
            )
          }
        >
          {pageNumber === "ellipsis" ? "..." : pageNumber}
        </button>
      ))}
      <button
        className="page-button"
        onClick={() => onPageChange(currentPage + 1)}
        disabled={currentPage === totalPages}
      >
        &raquo;
      </button>
    </div>
  );
};

export default PageButtons;
