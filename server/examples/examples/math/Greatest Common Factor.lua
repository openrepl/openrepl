-- gcf calculates the greatest common factor of x and y
-- implementation uses the Euclidean Algorithm
-- https://en.wikipedia.org/wiki/Euclidean_algorithm
function gcf(x, y)
    -- Greatest common factor of x and x is x.
    if x == y then
        return x
    end

    -- Swap order such that y > x.
    if x > y then
        y, x = x, y
    end

    -- gcf(x, y) = gcf(y-x, x) | y > x
    return gcf(y-x, x)
end

print("Greatest common factor of 55 and 30: " .. gcf(55, 30))
